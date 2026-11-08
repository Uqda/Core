package ipv6rwc

import (
	"encoding/binary"
	"encoding/hex"
	"io"
	"net/url"
	"testing"
	"time"

	"github.com/gologme/log"

	"github.com/Uqda/Core/src/address"
	"github.com/Uqda/Core/src/config"
	"github.com/Uqda/Core/src/core"
	"github.com/Uqda/Core/src/firewall"
)

func newTestCore(t *testing.T) *core.Core {
	t.Helper()
	cfg := config.GenerateConfig()
	if err := cfg.GenerateSelfSignedCertificate(); err != nil {
		t.Fatal(err)
	}
	logger := log.New(io.Discard, "", 0)
	c, err := core.New(cfg.Certificate, logger)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func packet(src, dst address.Address, proto uint8, sport, dport uint16) []byte {
	p := make([]byte, 40+20)
	p[0] = 0x60
	binary.BigEndian.PutUint16(p[4:6], 20)
	p[6] = proto
	p[7] = 64
	copy(p[8:24], src[:])
	copy(p[24:40], dst[:])
	binary.BigEndian.PutUint16(p[40:42], sport)
	binary.BigEndian.PutUint16(p[42:44], dport)
	return p
}

// TestFirewallFiltersRealOverlayTraffic runs two connected nodes and checks
// what node A's firewall delivers to its host from node B, over real
// encrypted sessions rather than synthetic calls.
func TestFirewallFiltersRealOverlayTraffic(t *testing.T) {
	coreA, coreB := newTestCore(t), newTestCore(t)
	defer coreA.Stop()
	defer coreB.Stop()

	listener, err := coreA.Listen(&url.URL{Scheme: "tcp", Host: "127.0.0.1:0"}, "")
	if err != nil {
		t.Fatal(err)
	}
	if err := coreB.CallPeer(&url.URL{Scheme: "tcp", Host: listener.Addr().String()}, ""); err != nil {
		t.Fatal(err)
	}

	rwcA, rwcB := NewReadWriteCloser(coreA), NewReadWriteCloser(coreB)
	fw, err := firewall.New(config.FirewallConfig{Enabled: true, Rules: []config.FirewallRule{
		{Sources: []string{hex.EncodeToString(coreB.PublicKey())}, Protocol: "tcp", Ports: []string{"22"}},
	}})
	if err != nil {
		t.Fatal(err)
	}
	rwcA.SetFirewall(fw)

	received := make(chan []byte, 256)
	go func() {
		buf := make([]byte, 65535)
		for {
			n, err := rwcA.Read(buf)
			if err != nil {
				return
			}
			received <- append([]byte(nil), buf[:n]...)
		}
	}()
	go func() { // drain B so its reads never block the session
		buf := make([]byte, 65535)
		for {
			if _, err := rwcB.Read(buf); err != nil {
				return
			}
		}
	}()

	addrA, addrB := rwcA.Address(), rwcB.Address()
	allowed := packet(addrB, addrA, 6, 40000, 22)
	denied := packet(addrB, addrA, 6, 40000, 23)
	reply := packet(addrB, addrA, 17, 9999, 5555)

	var gotAllowed, gotDenied, gotReply bool
	deadline := time.After(15 * time.Second)
	sentOutbound := false
	tick := time.NewTicker(100 * time.Millisecond)
	defer tick.Stop()
	for !gotAllowed || !gotReply {
		select {
		case <-tick.C:
			_, _ = rwcB.Write(allowed)
			_, _ = rwcB.Write(denied)
			if gotAllowed && !sentOutbound {
				// The path works; open an outbound UDP flow from A, then
				// keep sending B's reply until it arrives.
				_, _ = rwcA.Write(packet(addrA, addrB, 17, 5555, 9999))
				sentOutbound = true
			}
			if sentOutbound {
				_, _ = rwcB.Write(reply)
			}
		case pkt := <-received:
			dport := binary.BigEndian.Uint16(pkt[42:44])
			switch {
			case pkt[6] == 6 && dport == 22:
				gotAllowed = true
			case pkt[6] == 6 && dport == 23:
				gotDenied = true
			case pkt[6] == 17 && dport == 5555:
				gotReply = true
			}
		case <-deadline:
			t.Fatalf("timed out: allowed=%v reply=%v", gotAllowed, gotReply)
		}
	}
	// Give any stray denied packets a moment to show up.
	drain := time.After(500 * time.Millisecond)
	for done := false; !done; {
		select {
		case pkt := <-received:
			if pkt[6] == 6 && binary.BigEndian.Uint16(pkt[42:44]) == 23 {
				gotDenied = true
			}
		case <-drain:
			done = true
		}
	}
	if gotDenied {
		t.Fatal("a packet to a port no rule allows reached the host")
	}
	if s := fw.Status(); s.Dropped == 0 || s.Allowed == 0 {
		t.Fatalf("counters did not move: %+v", s)
	}
}
