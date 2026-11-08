package core

import (
	"bytes"
	"crypto/ed25519"
	"encoding/binary"
	"io"
	"net"
	"testing"
	"time"

	"github.com/Uqda/Core/src/config"
)

func newCoreWith(t *testing.T, opts ...SetupOption) *Core {
	t.Helper()
	cfg := config.GenerateConfig()
	requireNoError(t, cfg.GenerateSelfSignedCertificate())
	c, err := New(cfg.Certificate, GetLoggerWithPrefix("", false), opts...)
	requireNoError(t, err)
	t.Cleanup(c.Stop)
	return c
}

func listenTCP(t *testing.T, c *Core) string {
	t.Helper()
	l, err := c.Listen(mustParseURL(t, "tcp://127.0.0.1:0"), "")
	requireNoError(t, err)
	return l.Addr().String()
}

// readMeta reads one raw handshake metadata message.
func readMeta(t *testing.T, r io.Reader) []byte {
	t.Helper()
	head := make([]byte, 6)
	if _, err := io.ReadFull(r, head); err != nil {
		t.Fatalf("read metadata header: %v", err)
	}
	body := make([]byte, binary.BigEndian.Uint16(head[4:6]))
	if _, err := io.ReadFull(r, body); err != nil {
		t.Fatalf("read metadata body: %v", err)
	}
	return append(head, body...)
}

// stripNonce removes the fresh-handshake field from a raw metadata message.
// The signature covers only the public key, so the result still verifies.
func stripNonce(t *testing.T, raw []byte) []byte {
	t.Helper()
	body := raw[6 : len(raw)-ed25519.SignatureSize]
	sig := raw[len(raw)-ed25519.SignatureSize:]
	var kept []byte
	for len(body) >= 4 {
		op := binary.BigEndian.Uint16(body[:2])
		n := int(binary.BigEndian.Uint16(body[2:4]))
		field := body[:4+n]
		if op != metaUqdaFreshNonce {
			kept = append(kept, field...)
		}
		body = body[4+n:]
	}
	out := append([]byte("meta\x00\x00"), kept...)
	out = append(out, sig...)
	binary.BigEndian.PutUint16(out[4:6], uint16(len(out)-6))
	return out
}

func hasPeer(c *Core, key ed25519.PublicKey) (PeerInfo, bool) {
	for _, p := range c.GetPeers() {
		if bytes.Equal(p.Key, key) {
			return p, true
		}
	}
	return PeerInfo{}, false
}

// connClosedSoon reports whether the remote side closes conn within d.
func connClosedSoon(conn net.Conn, d time.Duration) bool {
	_ = conn.SetReadDeadline(time.Now().Add(d))
	buf := make([]byte, 256)
	for {
		if _, err := conn.Read(buf); err != nil {
			ne, ok := err.(net.Error)
			return !ok || !ne.Timeout()
		}
	}
}

func TestMetadataNonceRoundTripAndUnknownFieldsAreIgnored(t *testing.T) {
	pub, priv, _ := ed25519.GenerateKey(nil)
	m := versionGetBaseMetadata()
	m.publicKey = pub
	m.nonce = bytes.Repeat([]byte{7}, freshNonceSize)
	raw, err := m.encode(priv, nil)
	requireNoError(t, err)
	var got versionMetadata
	requireNoError(t, got.decode(bytes.NewReader(raw), nil))
	requireTrue(t, bytes.Equal(got.nonce, m.nonce))

	// A decoder must skip fields it does not know; that is how Yggdrasil
	// nodes treat the nonce. Insert an unknown field and decode again.
	body := raw[6 : len(raw)-ed25519.SignatureSize]
	extra := []byte{0x99, 0x99, 0x00, 0x02, 0xAB, 0xCD}
	withUnknown := append(append(append([]byte("meta\x00\x00"), body...), extra...), raw[len(raw)-ed25519.SignatureSize:]...)
	binary.BigEndian.PutUint16(withUnknown[4:6], uint16(len(withUnknown)-6))
	var again versionMetadata
	requireNoError(t, again.decode(bytes.NewReader(withUnknown), nil))
	requireTrue(t, again.check())

	// Without a nonce the encoding is exactly the Yggdrasil format.
	m.nonce = nil
	legacy, err := m.encode(priv, nil)
	requireNoError(t, err)
	requireTrue(t, !bytes.Contains(legacy, []byte{0x75, 0x71, 0x00, 0x20}))
}

func TestUqdaPeersUseTheFreshHandshake(t *testing.T) {
	a, b := newCoreWith(t, RequireFreshHandshake(true)), newCoreWith(t)
	requireNoError(t, b.CallPeer(mustParseURL(t, "tcp://"+listenTCP(t, a)), ""))
	ok := waitForCondition(5*time.Second, func() bool {
		pa, okA := hasPeer(a, b.PublicKey())
		pb, okB := hasPeer(b, a.PublicKey())
		return okA && okB && pa.Fresh && pb.Fresh
	})
	if !ok {
		t.Fatalf("Uqda nodes did not complete the fresh handshake: %+v / %+v", a.GetPeers(), b.GetPeers())
	}
}

// captureHandshake has node b dial a raw listener and records its metadata.
func captureHandshake(t *testing.T, b *Core) []byte {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	requireNoError(t, err)
	defer ln.Close()
	requireNoError(t, b.CallPeer(mustParseURL(t, "tcp://"+ln.Addr().String()), ""))
	requireNoError(t, ln.(*net.TCPListener).SetDeadline(time.Now().Add(5*time.Second)))
	conn, err := ln.Accept()
	requireNoError(t, err)
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(5 * time.Second))
	return readMeta(t, conn)
}

func TestReplayedHandshakeIsRejected(t *testing.T) {
	victim := newCoreWith(t)
	recorded := captureHandshake(t, victim)

	for _, tc := range []struct {
		name    string
		require bool
		replay  []byte
	}{
		{"replay with nonce, default mode", false, recorded},
		{"replay with nonce, fresh required", true, recorded},
		{"replay with nonce stripped, fresh required", true, stripNonce(t, recorded)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var opts []SetupOption
			if tc.require {
				opts = append(opts, RequireFreshHandshake(true))
			}
			target := newCoreWith(t, opts...)
			conn, err := net.Dial("tcp", listenTCP(t, target))
			requireNoError(t, err)
			defer conn.Close()
			_ = conn.SetDeadline(time.Now().Add(5 * time.Second))
			_, err = conn.Write(tc.replay)
			requireNoError(t, err)
			_ = readMeta(t, conn)
			// The attacker cannot sign the target's fresh nonce.
			_, _ = conn.Write(make([]byte, ed25519.SignatureSize))
			if !connClosedSoon(conn, 3*time.Second) {
				t.Fatal("target kept the replayed connection open")
			}
			if _, ok := hasPeer(target, victim.PublicKey()); ok {
				t.Fatal("replayed handshake produced a peering as the victim")
			}
		})
	}
}

// legacyHandshake connects to addr as a Yggdrasil-style node (no nonce).
func legacyHandshake(t *testing.T, addr string) (net.Conn, ed25519.PublicKey) {
	t.Helper()
	pub, priv, _ := ed25519.GenerateKey(nil)
	m := versionGetBaseMetadata()
	m.publicKey = pub
	raw, err := m.encode(priv, nil)
	requireNoError(t, err)
	conn, err := net.Dial("tcp", addr)
	requireNoError(t, err)
	_ = conn.SetDeadline(time.Now().Add(5 * time.Second))
	_, err = conn.Write(raw)
	requireNoError(t, err)
	_ = readMeta(t, conn)
	return conn, pub
}

func TestLegacyPeersAreAcceptedUnlessFreshIsRequired(t *testing.T) {
	open := newCoreWith(t)
	conn, key := legacyHandshake(t, listenTCP(t, open))
	defer conn.Close()
	if !waitForCondition(5*time.Second, func() bool { _, ok := hasPeer(open, key); return ok }) {
		t.Fatal("a Yggdrasil-style peer was not accepted in default mode")
	}
	if p, _ := hasPeer(open, key); p.Fresh {
		t.Fatal("a peer without a nonce was reported as fresh")
	}

	strict := newCoreWith(t, RequireFreshHandshake(true))
	conn2, key2 := legacyHandshake(t, listenTCP(t, strict))
	defer conn2.Close()
	if !connClosedSoon(conn2, 3*time.Second) {
		t.Fatal("strict node kept a legacy connection open")
	}
	if _, ok := hasPeer(strict, key2); ok {
		t.Fatal("strict node accepted a peer without a fresh handshake")
	}
}
