package encrypted

import (
	"crypto/ed25519"
	"encoding/hex"
	"testing"
)

func TestEdX25519(t *testing.T) {
	bsPub, bsPriv, err := ed25519.GenerateKey(nil)
	if err != nil {
		panic("key generation failed")
	}
	var ePub edPub
	var ePriv edPriv
	copy(ePub[:], bsPub)
	copy(ePriv[:], bsPriv)
	pub1, _ := ePub.toBox()
	priv1 := ePriv.toBox()
	pub2, priv2 := newBoxKeys()
	var encShared, decShared boxShared
	getShared(&encShared, pub1, &priv2)
	getShared(&decShared, &pub2, priv1)
	if encShared != decShared {
		panic("shared secret mismatch")
	}
}

func TestSessionInitPasswordAuth(t *testing.T) {
	// These fixtures are computed once, as they are for real PacketConns.
	auths := map[string]groupAuth{
		"":                newGroupAuth(""),
		"shared-password": newGroupAuth("shared-password"),
		"wrong-password":  newGroupAuth("wrong-password"),
	}
	legacy, err := hex.DecodeString("76ee79264fe680ae401d62d25e430d5a9300ccd9e5f2e6084b2c72aab317fbe8")
	if err != nil {
		t.Fatal(err)
	}
	legacyAuth := groupAuth{enabled: true}
	copy(legacyAuth.secret[:], legacy)
	auths["legacy"] = legacyAuth
	auths["modern"] = newGroupAuth("legacy-test-password-only")
	senderPub, senderPriv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("generate sender key: %v", err)
	}
	receiverPub, receiverPriv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("generate receiver key: %v", err)
	}

	var senderEd edPriv
	var senderEdPub edPub
	var receiverEdPriv edPriv
	var receiverEd edPub
	copy(senderEd[:], senderPriv)
	copy(senderEdPub[:], senderPub)
	copy(receiverEdPriv[:], receiverPriv)
	copy(receiverEd[:], receiverPub)

	current, _ := newBoxKeys()
	next, _ := newBoxKeys()

	init := newSessionInit(&current, &next, 9)
	tests := []struct {
		name             string
		senderPassword   string
		receiverPassword string
		wantOK           bool
	}{
		{
			name:             "both empty",
			senderPassword:   "",
			receiverPassword: "",
			wantOK:           true,
		},
		{
			name:             "both same password",
			senderPassword:   "shared-password",
			receiverPassword: "shared-password",
			wantOK:           true,
		},
		{
			name:             "sender password only",
			senderPassword:   "shared-password",
			receiverPassword: "",
			wantOK:           false,
		},
		{
			name:             "receiver password only",
			senderPassword:   "",
			receiverPassword: "shared-password",
			wantOK:           false,
		},
		{
			name:             "different passwords",
			senderPassword:   "shared-password",
			receiverPassword: "wrong-password",
			wantOK:           false,
		},
		{
			name: "legacy sender rejected", senderPassword: "legacy", receiverPassword: "modern", wantOK: false,
		},
		{
			name: "legacy receiver rejects modern", senderPassword: "modern", receiverPassword: "legacy", wantOK: false,
		},
	}

	receiverBox := receiverEdPriv.toBox()
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			data, err := init.encrypt(&senderEd, &receiverEd, auths[tc.senderPassword])
			if err != nil {
				t.Fatalf("encrypt handshake: %v", err)
			}

			var decoded sessionInit
			ok := decoded.decrypt(receiverBox, &senderEdPub, data, auths[tc.receiverPassword])
			if ok != tc.wantOK {
				t.Fatalf("decrypt = %v, want %v", ok, tc.wantOK)
			}
			if !ok {
				return
			}
			if decoded.current != init.current || decoded.next != init.next || decoded.keySeq != init.keySeq || decoded.seq != init.seq {
				t.Fatal("decoded handshake did not round-trip")
			}
		})
	}
}

func TestGroupAuthDerivation(t *testing.T) {
	public := newGroupAuth("")
	if public.enabled || public.preimage() != nil || public.secret != [32]byte{} {
		t.Fatal("public-mode authentication changed")
	}
	auth := newGroupAuth("uqda-test-vector-only")
	const want = "cb243c47777eff66f67cd925203dbf6624f913669308a9bedc953ff80895f6f8"
	if got := hex.EncodeToString(auth.secret[:]); !auth.enabled || got != want {
		t.Fatalf("Argon2id protocol vector = %s, want %s", got, want)
	}
}

func TestGroupAuthRejectsLegacySignature(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	var sender edPriv
	var receiver edPub
	copy(sender[:], priv)
	copy(receiver[:], pub)
	const password = "legacy-test-password-only"
	modern := newGroupAuth(password)
	// Frozen SHA-256 preimage from 26.0.1 for this test password only.
	legacy, err := hex.DecodeString("76ee79264fe680ae401d62d25e430d5a9300ccd9e5f2e6084b2c72aab317fbe8")
	if err != nil {
		t.Fatal(err)
	}
	message := []byte("session-authentication-test")
	if edCheck(message, edSign(message, &sender, legacy[:]), &receiver, modern.preimage()) {
		t.Fatal("accepted legacy private-group signature")
	}
	if edCheck(message, edSign(message, &sender, modern.preimage()), &receiver, legacy[:]) {
		t.Fatal("modern signature unexpectedly compatible with legacy group")
	}
	if !edCheck(message, edSign(message, &sender, modern.preimage()), &receiver, modern.preimage()) {
		t.Fatal("modern group signature rejected")
	}
}
