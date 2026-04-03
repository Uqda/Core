package core

import (
	"crypto/ed25519"
	"encoding/hex"
	"testing"
	"time"
)

func TestInviteTokenRoundTrip(t *testing.T) {
	pub, _, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	tok := &InviteToken{
		V:        1,
		Net:      "test-net",
		Peers:    []string{"tls://127.0.0.1:1"},
		Password: "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		OwnerKey: hex.EncodeToString(pub),
		Expires:  time.Now().Add(time.Hour).Unix(),
		Admin:    "tcp://127.0.0.1:9001",
	}
	s, err := EncodeInviteToken(tok)
	if err != nil {
		t.Fatal(err)
	}
	dec, err := DecodeInviteToken(s)
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateInviteToken(dec); err != nil {
		t.Fatal(err)
	}
	if dec.Net != tok.Net || dec.OwnerKey != tok.OwnerKey {
		t.Fatalf("mismatch: %+v vs %+v", dec, tok)
	}
	if err := ValidateInviteOwner(dec, pub); err != nil {
		t.Fatal(err)
	}
}
