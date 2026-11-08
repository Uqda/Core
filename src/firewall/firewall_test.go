package firewall

import (
	"crypto/ed25519"
	"encoding/binary"
	"encoding/hex"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/Uqda/Core/src/address"
	"github.com/Uqda/Core/src/config"
)

type node struct {
	pub  ed25519.PublicKey
	addr address.Address
}

func newNode(t *testing.T) node {
	t.Helper()
	pub, _, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	return node{pub: pub, addr: *address.AddrForKey(pub)}
}

func ipv6(src, dst address.Address, nextHeader uint8, payload []byte) []byte {
	p := make([]byte, 40+len(payload))
	p[0] = 0x60
	binary.BigEndian.PutUint16(p[4:6], uint16(len(payload)))
	p[6] = nextHeader
	p[7] = 64
	copy(p[8:24], src[:])
	copy(p[24:40], dst[:])
	copy(p[40:], payload)
	return p
}

func ports(src, dst uint16, extra int) []byte {
	b := make([]byte, 4+extra)
	binary.BigEndian.PutUint16(b[0:2], src)
	binary.BigEndian.PutUint16(b[2:4], dst)
	return b
}

func tcp(src, dst address.Address, sport, dport uint16) []byte {
	return ipv6(src, dst, protoTCP, ports(sport, dport, 16))
}

func udp(src, dst address.Address, sport, dport uint16) []byte {
	return ipv6(src, dst, protoUDP, ports(sport, dport, 4))
}

func icmp(src, dst address.Address, typ uint8, id uint16) []byte {
	b := make([]byte, 8)
	b[0] = typ
	binary.BigEndian.PutUint16(b[4:6], id)
	return ipv6(src, dst, protoICMPv6, b)
}

func mustNew(t *testing.T, cfg config.FirewallConfig) *Firewall {
	t.Helper()
	cfg.Enabled = true
	f, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	return f
}

func TestDisabledFirewallIsNilAndAllowsEverything(t *testing.T) {
	f, err := New(config.FirewallConfig{Enabled: false})
	if err != nil || f != nil {
		t.Fatalf("disabled firewall: got %v, %v", f, err)
	}
	a, b := newNode(t), newNode(t)
	if !f.Inbound(tcp(b.addr, a.addr, 1000, 22), b.pub) {
		t.Fatal("nil firewall must allow inbound traffic")
	}
	f.Outbound(tcp(a.addr, b.addr, 22, 1000)) // must not panic
	if f.Status().Enabled {
		t.Fatal("nil firewall must report disabled")
	}
}

func TestDefaultDenyDropsUnsolicitedTCPAndUDP(t *testing.T) {
	f := mustNew(t, config.FirewallConfig{})
	a, b := newNode(t), newNode(t)
	if f.Inbound(tcp(b.addr, a.addr, 40000, 22), b.pub) {
		t.Fatal("unsolicited TCP was allowed")
	}
	if f.Inbound(udp(b.addr, a.addr, 40000, 53), b.pub) {
		t.Fatal("unsolicited UDP was allowed")
	}
	if s := f.Status(); s.Dropped != 2 || s.Allowed != 0 {
		t.Fatalf("counters: %+v", s)
	}
}

func TestRepliesToOutboundFlowsAreAllowed(t *testing.T) {
	f := mustNew(t, config.FirewallConfig{})
	a, b, c := newNode(t), newNode(t), newNode(t)

	f.Outbound(tcp(a.addr, b.addr, 50000, 443))
	if !f.Inbound(tcp(b.addr, a.addr, 443, 50000), b.pub) {
		t.Fatal("TCP reply was dropped")
	}
	// Same ports from a different node, or a different port, are not replies.
	if f.Inbound(tcp(c.addr, a.addr, 443, 50000), c.pub) {
		t.Fatal("a third node reused the flow")
	}
	if f.Inbound(tcp(b.addr, a.addr, 443, 50001), b.pub) {
		t.Fatal("a different local port matched the flow")
	}

	f.Outbound(udp(a.addr, b.addr, 50002, 53))
	if !f.Inbound(udp(b.addr, a.addr, 53, 50002), b.pub) {
		t.Fatal("UDP reply was dropped")
	}
}

func TestFlowsExpire(t *testing.T) {
	f := mustNew(t, config.FirewallConfig{})
	now := time.Unix(1_000_000, 0)
	f.now = func() time.Time { return now }
	a, b := newNode(t), newNode(t)

	f.Outbound(udp(a.addr, b.addr, 5000, 6000))
	now = now.Add(udpFlowTimeout - time.Second)
	if !f.Inbound(udp(b.addr, a.addr, 6000, 5000), b.pub) {
		t.Fatal("reply before expiry was dropped")
	}
	// The reply refreshed the flow; go past a full timeout with no traffic.
	now = now.Add(udpFlowTimeout + time.Second)
	if f.Inbound(udp(b.addr, a.addr, 6000, 5000), b.pub) {
		t.Fatal("reply after expiry was allowed")
	}
}

func TestPingAndICMPErrors(t *testing.T) {
	a, b := newNode(t), newNode(t)

	blocked := mustNew(t, config.FirewallConfig{AllowPing: false})
	if blocked.Inbound(icmp(b.addr, a.addr, 128, 7), b.pub) {
		t.Fatal("echo request allowed with AllowPing off")
	}
	if !blocked.Inbound(icmp(b.addr, a.addr, 2, 0), b.pub) {
		t.Fatal("Packet Too Big must always be allowed for path MTU discovery")
	}
	if !blocked.Inbound(icmp(b.addr, a.addr, 1, 0), b.pub) {
		t.Fatal("Destination Unreachable must always be allowed")
	}
	if blocked.Inbound(icmp(b.addr, a.addr, 129, 7), b.pub) {
		t.Fatal("unsolicited echo reply was allowed")
	}
	blocked.Outbound(icmp(a.addr, b.addr, 128, 7))
	if !blocked.Inbound(icmp(b.addr, a.addr, 129, 7), b.pub) {
		t.Fatal("echo reply to our own ping was dropped")
	}
	if blocked.Inbound(icmp(b.addr, a.addr, 129, 8), b.pub) {
		t.Fatal("echo reply with another identifier was allowed")
	}

	open := mustNew(t, config.FirewallConfig{AllowPing: true})
	if !open.Inbound(icmp(b.addr, a.addr, 128, 7), b.pub) {
		t.Fatal("echo request dropped with AllowPing on")
	}
}

func TestRulesBySourceProtocolAndPort(t *testing.T) {
	a, friend, stranger := newNode(t), newNode(t), newNode(t)
	sub := address.SubnetForKey(friend.pub)
	subnetText := net.IP(append(sub[:], make([]byte, 8)...)).String() + "/64"

	cases := []struct {
		name   string
		source string
	}{
		{"public key", hex.EncodeToString(friend.pub)},
		{"address", net.IP(friend.addr[:]).String()},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := mustNew(t, config.FirewallConfig{Rules: []config.FirewallRule{
				{Sources: []string{tc.source}, Protocol: "tcp", Ports: []string{"22", "8000-8100"}},
			}})
			if !f.Inbound(tcp(friend.addr, a.addr, 40000, 22), friend.pub) {
				t.Fatal("allowed source was dropped on port 22")
			}
			if !f.Inbound(tcp(friend.addr, a.addr, 40000, 8050), friend.pub) {
				t.Fatal("allowed source was dropped inside the range")
			}
			if f.Inbound(tcp(friend.addr, a.addr, 40000, 8101), friend.pub) {
				t.Fatal("port outside the range was allowed")
			}
			if f.Inbound(udp(friend.addr, a.addr, 40000, 22), friend.pub) {
				t.Fatal("UDP matched a TCP rule")
			}
			if f.Inbound(tcp(stranger.addr, a.addr, 40000, 22), stranger.pub) {
				t.Fatal("another node matched the rule")
			}
		})
	}

	t.Run("subnet", func(t *testing.T) {
		f := mustNew(t, config.FirewallConfig{Rules: []config.FirewallRule{
			{Sources: []string{subnetText}, Protocol: "any"},
		}})
		var hostInSubnet address.Address
		copy(hostInSubnet[:], sub[:])
		hostInSubnet[15] = 9
		if !f.Inbound(udp(hostInSubnet, a.addr, 1, 2), friend.pub) {
			t.Fatal("host inside the allowed /64 was dropped")
		}
		if f.Inbound(udp(stranger.addr, a.addr, 1, 2), stranger.pub) {
			t.Fatal("host outside the /64 was allowed")
		}
	})

	t.Run("any source", func(t *testing.T) {
		f := mustNew(t, config.FirewallConfig{Rules: []config.FirewallRule{
			{Sources: []string{"*"}, Protocol: "tcp", Ports: []string{"443"}},
		}})
		if !f.Inbound(tcp(stranger.addr, a.addr, 1, 443), stranger.pub) {
			t.Fatal("public port was dropped")
		}
		if f.Inbound(tcp(stranger.addr, a.addr, 1, 444), stranger.pub) {
			t.Fatal("non-public port was allowed")
		}
	})
}

func TestLaterFragmentsNeedAFlowOrAnAllPortsRule(t *testing.T) {
	f := mustNew(t, config.FirewallConfig{})
	a, b := newNode(t), newNode(t)
	frag := make([]byte, 8)
	frag[0] = protoUDP
	binary.BigEndian.PutUint16(frag[2:4], 100<<3) // offset 100, a later fragment
	pkt := ipv6(b.addr, a.addr, 44, append(frag, make([]byte, 16)...))

	if f.Inbound(pkt, b.pub) {
		t.Fatal("unsolicited later fragment was allowed")
	}
	f.Outbound(udp(a.addr, b.addr, 5000, 6000))
	if !f.Inbound(pkt, b.pub) {
		t.Fatal("later fragment from a node we are talking to was dropped")
	}
}

func TestExtensionHeadersAreWalked(t *testing.T) {
	f := mustNew(t, config.FirewallConfig{Rules: []config.FirewallRule{
		{Sources: []string{"*"}, Protocol: "tcp", Ports: []string{"22"}},
	}})
	a, b := newNode(t), newNode(t)
	// Destination options header (8 bytes) followed by TCP to port 22.
	opts := make([]byte, 8)
	opts[0] = protoTCP
	pkt := ipv6(b.addr, a.addr, 60, append(opts, ports(1000, 22, 16)...))
	if !f.Inbound(pkt, b.pub) {
		t.Fatal("TCP behind a destination options header was not recognised")
	}
	opts[0] = protoTCP
	pkt = ipv6(b.addr, a.addr, 60, append(opts, ports(1000, 23, 16)...))
	if f.Inbound(pkt, b.pub) {
		t.Fatal("extension header was used to bypass the port check")
	}
}

func TestMalformedPacketsAreDropped(t *testing.T) {
	f := mustNew(t, config.FirewallConfig{Rules: []config.FirewallRule{{Sources: []string{"*"}}}})
	a, b := newNode(t), newNode(t)
	full := tcp(b.addr, a.addr, 1, 2)
	for _, pkt := range [][]byte{nil, full[:20], full[:41]} {
		if f.Inbound(pkt, b.pub) {
			t.Fatalf("truncated packet of %d bytes was allowed", len(pkt))
		}
	}
	// A header length that runs past the end of the packet.
	bad := ipv6(b.addr, a.addr, 60, []byte{protoTCP, 200, 0, 0})
	if f.Inbound(bad, b.pub) {
		t.Fatal("packet with an overlong extension header was allowed")
	}
}

func TestFlowTableIsBounded(t *testing.T) {
	f := mustNew(t, config.FirewallConfig{})
	a, b := newNode(t), newNode(t)
	for i := 0; i < maxFlows+10; i++ {
		f.Outbound(udp(a.addr, b.addr, uint16(i%65000+1), uint16(i/65000+1)))
	}
	if s := f.Status(); s.ActiveFlows > maxFlows || s.Untracked == 0 {
		t.Fatalf("flow table not bounded: %+v", s)
	}
}

func TestValidateReportsEveryBadRule(t *testing.T) {
	cfg := config.FirewallConfig{Enabled: true, Rules: []config.FirewallRule{
		{Sources: nil},
		{Sources: []string{"10.0.0.1"}},
		{Sources: []string{"*"}, Protocol: "sctp"},
		{Sources: []string{"*"}, Protocol: "tcp", Ports: []string{"0"}},
		{Sources: []string{"*"}, Protocol: "tcp", Ports: []string{"90-80"}},
		{Sources: []string{"*"}, Protocol: "icmp", Ports: []string{"22"}},
		{Sources: []string{"2001:db8::1"}},
		{Sources: []string{"*"}, Protocol: "tcp", Ports: []string{"22"}}, // valid
	}}
	problems := Validate(cfg)
	if len(problems) != 7 {
		t.Fatalf("expected 7 problems, got %d: %s", len(problems), strings.Join(problems, "; "))
	}
	if _, err := New(cfg); err == nil {
		t.Fatal("New accepted an invalid configuration")
	}
}

func TestStatusDescribesRules(t *testing.T) {
	f := mustNew(t, config.FirewallConfig{AllowPing: true, Rules: []config.FirewallRule{
		{Sources: []string{"*"}, Protocol: "tcp", Ports: []string{"22"}, Comment: "SSH"},
	}})
	s := f.Status()
	if !s.Enabled || !s.AllowPing || len(s.Rules) != 1 || !strings.Contains(s.Rules[0], "SSH") {
		t.Fatalf("unexpected status: %+v", s)
	}
}

func FuzzParse(f *testing.F) {
	a, b := address.Address{2}, address.Address{2, 1}
	f.Add(tcp(a, b, 1, 2))
	f.Add(icmp(a, b, 128, 1))
	f.Add(ipv6(a, b, 60, []byte{protoTCP, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 2}))
	f.Add(ipv6(a, b, 44, []byte{protoUDP, 0, 0, 8, 0, 0, 0, 0}))
	fw, _ := New(config.FirewallConfig{Enabled: true, Rules: []config.FirewallRule{{Sources: []string{"*"}, Protocol: "tcp", Ports: []string{"1-100"}}}})
	key := make(ed25519.PublicKey, ed25519.PublicKeySize)
	f.Fuzz(func(t *testing.T, pkt []byte) {
		_ = parse(pkt)
		if len(pkt) >= 40 {
			fw.Outbound(pkt)
			_ = fw.Inbound(pkt, key)
		}
	})
}
