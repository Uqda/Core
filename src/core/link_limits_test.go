package core

import (
	"errors"
	"io"
	"net"
	"testing"
	"time"

	"github.com/Uqda/Core/src/config"
)

func TestHandshakeLimiterBounds(t *testing.T) {
	h := newHandshakeLimiter(3, 2)

	relA1, ok := h.acquire("a")
	requireTrue(t, ok)
	relA2, ok := h.acquire("a")
	requireTrue(t, ok)
	if _, ok := h.acquire("a"); ok {
		t.Fatal("per-source bound was not enforced")
	}
	relB1, ok := h.acquire("b")
	requireTrue(t, ok)
	if _, ok := h.acquire("c"); ok {
		t.Fatal("global bound was not enforced")
	}

	// Release is idempotent: a double release must not free an extra slot.
	relA1()
	relA1()
	relA3, ok := h.acquire("a")
	requireTrue(t, ok)
	if _, ok := h.acquire("c"); ok {
		t.Fatal("double release freed more than one slot")
	}

	relA2()
	relA3()
	relB1()
	h.mu.Lock()
	defer h.mu.Unlock()
	requireEqual(t, h.total, 0)
	requireEqual(t, len(h.perSource), 0)
}

func TestHandshakeSourceKeyGroupsIPv6By64(t *testing.T) {
	a := handshakeSourceKey(&net.TCPAddr{IP: net.ParseIP("2001:db8:1:2::1"), Port: 1})
	b := handshakeSourceKey(&net.TCPAddr{IP: net.ParseIP("2001:db8:1:2:ffff::9"), Port: 2})
	c := handshakeSourceKey(&net.TCPAddr{IP: net.ParseIP("2001:db8:1:3::1"), Port: 1})
	requireEqual(t, a, b)
	if a == c {
		t.Fatal("different /64 prefixes must not share a key")
	}
	requireEqual(t, handshakeSourceKey(&net.TCPAddr{IP: net.ParseIP("192.0.2.7"), Port: 9}), "192.0.2.7")
	requireEqual(t, handshakeSourceKey(&net.TCPAddr{IP: net.ParseIP("::ffff:192.0.2.7"), Port: 9}), "192.0.2.7")
	requireEqual(t, handshakeSourceKey(&net.UnixAddr{Name: "/tmp/x", Net: "unix"}), "unix")
	requireEqual(t, handshakeSourceKey(nil), "unknown")
}

// TestInboundHandshakeFloodIsBounded opens more idle connections than one
// source may hold in the handshake, checks the excess is closed at once
// (rather than after the six-second handshake deadline), and checks that a
// real peer can connect once the idle connections go away.
func TestInboundHandshakeFloodIsBounded(t *testing.T) {
	nodeA, nodeB := CreateAndConnectTwoUnconnected(t)
	defer nodeA.Stop()
	defer nodeB.Stop()

	listener, err := nodeA.Listen(mustParseURL(t, "tcp://127.0.0.1:0"), "")
	requireNoError(t, err)
	addr := listener.Addr().String()

	idle := make([]net.Conn, 0, maxPendingHandshakesPerSource)
	defer func() {
		for _, c := range idle {
			_ = c.Close()
		}
	}()
	for i := 0; i < maxPendingHandshakesPerSource; i++ {
		c, err := net.Dial("tcp", addr)
		requireNoError(t, err)
		idle = append(idle, c)
		// Accepted connections receive our handshake metadata first.
		requireNoError(t, c.SetReadDeadline(time.Now().Add(3*time.Second)))
		buf := make([]byte, 4)
		if _, err := io.ReadFull(c, buf); err != nil {
			t.Fatalf("idle connection %d was not accepted: %v", i, err)
		}
	}

	excess, err := net.Dial("tcp", addr)
	requireNoError(t, err)
	defer excess.Close()
	requireNoError(t, excess.SetReadDeadline(time.Now().Add(3*time.Second)))
	n, err := excess.Read(make([]byte, 64))
	var netErr net.Error
	switch {
	case errors.As(err, &netErr) && netErr.Timeout():
		t.Fatal("excess connection was held open instead of being rejected")
	case n != 0 || err == nil:
		t.Fatalf("excess connection received %d handshake bytes; expected rejection", n)
	}

	for _, c := range idle {
		_ = c.Close()
	}
	idle = nil

	// The released slots must allow a legitimate peering from the same source.
	requireNoError(t, nodeB.CallPeer(mustParseURL(t, "tcp://"+addr), ""))
	if !waitForCondition(5*time.Second, func() bool { return len(nodeA.GetPeers()) == 1 }) {
		t.Fatal("legitimate peer could not connect after the flood ended")
	}
}

// Ensure the default Core wiring creates the limiter with the documented bounds.
func TestHandshakeLimiterDefaults(t *testing.T) {
	cfg := config.GenerateConfig()
	requireNoError(t, cfg.GenerateSelfSignedCertificate())
	c, err := New(cfg.Certificate, GetLoggerWithPrefix("", false))
	requireNoError(t, err)
	defer c.Stop()
	requireEqual(t, c.links.handshakes.maxTotal, maxPendingHandshakes)
	requireEqual(t, c.links.handshakes.maxSource, maxPendingHandshakesPerSource)
}
