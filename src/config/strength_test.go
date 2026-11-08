package config

import (
	"crypto/ed25519"
	"testing"

	"github.com/Uqda/Core/src/address"
)

func TestKeyStrength(t *testing.T) {
	key := make(ed25519.PublicKey, ed25519.PublicKeySize)
	key[0], key[1] = 0x00, 0x1f // 8 + 3 leading zero bits
	if got := KeyStrength(key); got != 11 {
		t.Fatalf("KeyStrength = %d, want 11", got)
	}
	key[0] = 0x80
	if got := KeyStrength(key); got != 0 {
		t.Fatalf("KeyStrength = %d, want 0", got)
	}
}

func TestNewPrivateKeyWithStrength(t *testing.T) {
	cfg := GenerateConfig()
	if err := cfg.NewPrivateKeyWithStrength(10); err != nil {
		t.Fatal(err)
	}
	pub := ed25519.PrivateKey(cfg.PrivateKey).Public().(ed25519.PublicKey)
	if KeyStrength(pub) < 10 {
		t.Fatalf("key strength %d below 10", KeyStrength(pub))
	}
	// The address encodes the strength in its second byte, so a stronger
	// key gives a visibly higher address (for example 20a: instead of 200:).
	if addr := address.AddrForKey(pub); addr == nil || int(addr[1]) < 10 {
		t.Fatalf("address does not reflect the key strength: %v", addr)
	}
	if err := cfg.NewPrivateKeyWithStrength(MaxKeyStrength + 1); err == nil {
		t.Fatal("strength above the maximum must be rejected")
	}
}
