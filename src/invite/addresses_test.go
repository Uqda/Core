package invite

import (
	"net"
	"reflect"
	"testing"
)

func TestInviteAddressesSkipVirtualAdaptersAndPreferLAN(t *testing.T) {
	c := func(iface, ip string) candidateAddr { return candidateAddr{iface: iface, ip: net.ParseIP(ip)} }
	// Shaped like the Windows machine that reported the problem.
	got := pickInviteAddresses([]candidateAddr{
		c("vEthernet (WSL (Hyper-V firewall))", "172.24.96.1"),
		c("vEthernet (Default Switch)", "172.21.96.1"),
		c("WLAN", "fdc5:f709:58f6:4cd0:49ee:ddb9:7296:32fa"),
		c("WLAN", "fe80::1"),
		c("WLAN", "192.168.0.38"),
		c("Uqda", "200:74e4:40f2:6946:513f:33ba:903d:aa27"),
		c("docker0", "172.17.0.1"),
		c("eth0", "203.0.113.9"),
		c("eth0", "2001:db8::9"),
	}, 63371)
	want := []string{
		"192.168.0.38:63371",
		"[fdc5:f709:58f6:4cd0:49ee:ddb9:7296:32fa]:63371",
		"203.0.113.9:63371",
		"[2001:db8::9]:63371",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v\nwant %v", got, want)
	}
}

func TestInviteAddressesAreCapped(t *testing.T) {
	var cands []candidateAddr
	for i := 1; i <= 10; i++ {
		cands = append(cands, candidateAddr{iface: "eth0", ip: net.IPv4(192, 168, 1, byte(i))})
	}
	if got := pickInviteAddresses(cands, 1); len(got) != maxInviteAddresses {
		t.Fatalf("expected %d addresses, got %d", maxInviteAddresses, len(got))
	}
}

func TestVirtualInterfaceNames(t *testing.T) {
	for name, want := range map[string]bool{
		"vEthernet (WSL)": true, "docker0": true, "br-1a2b": true, "veth12": true,
		"utun3": true, "uqda0": true, "VirtualBox Host-Only Network": true,
		"WLAN": false, "Ethernet": false, "eth0": false, "en0": false, "wlp2s0": false,
	} {
		if got := virtualInterface(name); got != want {
			t.Errorf("virtualInterface(%q) = %v, want %v", name, got, want)
		}
	}
}
