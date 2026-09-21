package tun

import (
	"encoding/binary"
	"net"
	"testing"
)

func TestNativeIPv6WordsPreserveCompressedAddressBytes(t *testing.T) {
	const cidr = "200:db8::1234/64"
	words, err := nativeIPv6Words(cidr)
	if err != nil {
		t.Fatal(err)
	}
	var got [16]byte
	for i, word := range words {
		binary.NativeEndian.PutUint16(got[i*2:i*2+2], word)
	}
	want, _, _ := net.ParseCIDR(cidr)
	if !net.IP(got[:]).Equal(want) {
		t.Fatalf("converted address is %s, want %s", net.IP(got[:]), want)
	}
}

func TestNativeIPv6WordsRejectsInvalidAddress(t *testing.T) {
	for _, address := range []string{"not-an-address", "192.0.2.1/24"} {
		if _, err := nativeIPv6Words(address); err == nil {
			t.Fatalf("accepted invalid IPv6 TUN address %q", address)
		}
	}
}
