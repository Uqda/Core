package core

import (
	"bytes"
	"crypto/ed25519"
	"encoding/binary"
	"reflect"
	"testing"

	"golang.org/x/crypto/blake2b"
)

func TestVersionPasswordAuth(t *testing.T) {
	for _, tt := range []struct {
		password1 []byte // The password on node 1
		password2 []byte // The password on node 2
		allowed   bool   // Should the connection have been allowed?
	}{
		{nil, nil, true},                      // Allow:  No passwords (both nil)
		{nil, []byte(""), true},               // Allow:  No passwords (mixed nil and empty string)
		{nil, []byte("foo"), false},           // Reject: One node has a password, the other doesn't
		{[]byte("foo"), []byte(""), false},    // Reject: One node has a password, the other doesn't
		{[]byte("foo"), []byte("foo"), true},  // Allow:  Same password
		{[]byte("foo"), []byte("bar"), false}, // Reject: Different passwords
	} {
		pk1, sk1, err := ed25519.GenerateKey(nil)
		if err != nil {
			t.Fatalf("Node 1 failed to generate key: %s", err)
		}

		metadata1 := &version_metadata{
			publicKey: pk1,
		}
		encoded, err := metadata1.encode(sk1, tt.password1)
		if err != nil {
			t.Fatalf("Node 1 failed to encode metadata: %s", err)
		}

		var decoded version_metadata
		if allowed := decoded.decode(bytes.NewBuffer(encoded), tt.password2) == nil; allowed != tt.allowed {
			t.Fatalf("Permutation %q -> %q should have been %v but was %v", tt.password1, tt.password2, tt.allowed, allowed)
		}
	}
}

func TestVersionRoundtrip(t *testing.T) {
	for _, password := range [][]byte{
		nil, []byte(""), []byte("foo"),
	} {
		for _, test := range []*version_metadata{
			{majorVer: 1},
			{majorVer: 256},
			{majorVer: 2, minorVer: 4},
			{majorVer: 2, minorVer: 257},
			{majorVer: 258, minorVer: 259},
			{majorVer: 3, minorVer: 5, priority: 6},
			{majorVer: 260, minorVer: 261, priority: 7},
		} {
			// Generate a random public key for each time, since it is
			// a required field.
			pk, sk, err := ed25519.GenerateKey(nil)
			if err != nil {
				t.Fatal(err)
			}

			test.publicKey = pk
			meta, err := test.encode(sk, password)
			if err != nil {
				t.Fatal(err)
			}
			encoded := bytes.NewBuffer(meta)
			decoded := &version_metadata{}
			if err := decoded.decode(encoded, password); err != nil {
				t.Fatalf("failed to decode: %s", err)
			}
			if !reflect.DeepEqual(test, decoded) {
				t.Fatalf("round-trip failed\nwant: %+v\n got: %+v", test, decoded)
			}
		}
	}
}

func TestVersionSignatureCoversAllMetadata(t *testing.T) {
	pk, sk, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	metadata := &version_metadata{
		majorVer:  ProtocolVersionMajor,
		minorVer:  ProtocolVersionMinor,
		publicKey: pk,
		priority:  7,
	}
	encoded, err := metadata.encode(sk, []byte("shared secret"))
	if err != nil {
		t.Fatal(err)
	}

	// The priority is the byte immediately before the nonce TLV. Changing it
	// must invalidate the signature rather than silently changing routing input.
	for i := 6; i+5 < len(encoded)-ed25519.SignatureSize; {
		op := binary.BigEndian.Uint16(encoded[i : i+2])
		fieldLen := int(binary.BigEndian.Uint16(encoded[i+2 : i+4]))
		if op == metaPriority {
			encoded[i+4] ^= 0xff
			break
		}
		i += 4 + fieldLen
	}
	var decoded version_metadata
	if err := decoded.decode(bytes.NewReader(encoded), []byte("shared secret")); err != ErrHandshakeInvalidSignature {
		t.Fatalf("tampered metadata returned %v, want %v", err, ErrHandshakeInvalidSignature)
	}
}

func TestConfirmationBindsBothHellos(t *testing.T) {
	pk1, sk1, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	pk2, sk2, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	password := []byte("shared secret")
	local := &version_metadata{majorVer: ProtocolVersionMajor, minorVer: ProtocolVersionMinor, publicKey: pk1}
	remote := &version_metadata{majorVer: ProtocolVersionMajor, minorVer: ProtocolVersionMinor, publicKey: pk2}
	if _, err = local.encode(sk1, password); err != nil {
		t.Fatal(err)
	}
	if _, err = remote.encode(sk2, password); err != nil {
		t.Fatal(err)
	}
	confirmation, err := encodeConfirmation(sk1, password, local, remote)
	if err != nil {
		t.Fatal(err)
	}
	if err = decodeConfirmation(bytes.NewReader(confirmation), pk1, password, remote, local); err != nil {
		t.Fatalf("valid confirmation rejected: %v", err)
	}

	// A fresh hello changes the transcript, so a captured confirmation cannot
	// be replayed into the new connection.
	freshRemote := &version_metadata{majorVer: ProtocolVersionMajor, minorVer: ProtocolVersionMinor, publicKey: pk2}
	if _, err = freshRemote.encode(sk2, password); err != nil {
		t.Fatal(err)
	}
	if err = decodeConfirmation(bytes.NewReader(confirmation), pk1, password, freshRemote, local); err != ErrHandshakeInvalidConfirmation {
		t.Fatalf("replayed confirmation returned %v, want %v", err, ErrHandshakeInvalidConfirmation)
	}
}

func TestHandshakeCompatibilityWithLegacy05(t *testing.T) {
	pk, sk, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	password := []byte("shared secret")

	newMetadata := &version_metadata{
		majorVer:  ProtocolVersionMajor,
		minorVer:  ProtocolVersionMinor,
		publicKey: pk,
		priority:  3,
	}
	newWire, err := newMetadata.encode(sk, password)
	if err != nil {
		t.Fatal(err)
	}
	if err = decodeLegacy05ForTest(newWire, password); err != nil {
		t.Fatalf("legacy 0.5 decoder rejected extended hello: %v", err)
	}

	legacyWire := encodeLegacy05ForTest(t, sk, pk, password, 4)
	var decoded version_metadata
	if err = decoded.decode(bytes.NewReader(legacyWire), password); err != nil {
		t.Fatalf("new decoder rejected legacy 0.5 hello: %v", err)
	}
	if decoded.supportsHandshakeConfirmation() {
		t.Fatal("legacy hello unexpectedly advertised secure confirmation")
	}
}

func TestVersionDecodeRejectsMalformedFieldLengths(t *testing.T) {
	password := []byte("pw")
	for _, tt := range []struct {
		name  string
		op    uint16
		field []byte
	}{
		{name: "major short", op: metaVersionMajor, field: []byte{1}},
		{name: "minor short", op: metaVersionMinor, field: []byte{1}},
		{name: "public key short", op: metaPublicKey, field: []byte{1}},
		{name: "priority empty", op: metaPriority, field: nil},
	} {
		t.Run(tt.name, func(t *testing.T) {
			msg := malformedVersionHandshake(t, tt.op, tt.field, password)
			var decoded version_metadata
			if err := decoded.decode(bytes.NewReader(msg), password); err != ErrHandshakeInvalidLength {
				t.Fatalf("expected %q, got %v", ErrHandshakeInvalidLength, err)
			}
		})
	}
}

func TestVersionDecodeRejectsTrailingBytes(t *testing.T) {
	password := []byte("pw")
	pk, sk, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}

	hasher, err := blake2b.New512(password)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = hasher.Write(pk); err != nil {
		t.Fatal(err)
	}
	sig := ed25519.Sign(sk, hasher.Sum(nil))

	body := append([]byte{1, 2, 3}, sig...)
	msg := append([]byte{'m', 'e', 't', 'a', 0, 0}, body...)
	binary.BigEndian.PutUint16(msg[4:6], uint16(len(body)))
	var decoded version_metadata
	if err := decoded.decode(bytes.NewReader(msg), password); err != ErrHandshakeInvalidLength {
		t.Fatalf("expected %q, got %v", ErrHandshakeInvalidLength, err)
	}
}

func malformedVersionHandshake(t *testing.T, op uint16, field []byte, password []byte) []byte {
	t.Helper()

	pk, sk, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}

	hasher, err := blake2b.New512(password)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = hasher.Write(pk); err != nil {
		t.Fatal(err)
	}
	sig := ed25519.Sign(sk, hasher.Sum(nil))

	body := make([]byte, 0, 4+len(field)+len(sig))
	body = binary.BigEndian.AppendUint16(body, op)
	body = binary.BigEndian.AppendUint16(body, uint16(len(field)))
	body = append(body, field...)
	body = append(body, sig...)

	msg := append([]byte{'m', 'e', 't', 'a', 0, 0}, body...)
	binary.BigEndian.PutUint16(msg[4:6], uint16(len(body)))
	return msg
}

func encodeLegacy05ForTest(t *testing.T, privateKey ed25519.PrivateKey, publicKey ed25519.PublicKey, password []byte, priority uint8) []byte {
	t.Helper()
	body := make([]byte, 0, 128)
	body = binary.BigEndian.AppendUint16(body, metaVersionMajor)
	body = binary.BigEndian.AppendUint16(body, 2)
	body = binary.BigEndian.AppendUint16(body, ProtocolVersionMajor)
	body = binary.BigEndian.AppendUint16(body, metaVersionMinor)
	body = binary.BigEndian.AppendUint16(body, 2)
	body = binary.BigEndian.AppendUint16(body, ProtocolVersionMinor)
	body = binary.BigEndian.AppendUint16(body, metaPublicKey)
	body = binary.BigEndian.AppendUint16(body, ed25519.PublicKeySize)
	body = append(body, publicKey...)
	body = binary.BigEndian.AppendUint16(body, metaPriority)
	body = binary.BigEndian.AppendUint16(body, 1)
	body = append(body, priority)
	hash, err := legacyHandshakeHash(password, publicKey)
	if err != nil {
		t.Fatal(err)
	}
	body = append(body, ed25519.Sign(privateKey, hash)...)
	message := append([]byte{'m', 'e', 't', 'a', 0, 0}, body...)
	binary.BigEndian.PutUint16(message[4:6], uint16(len(body)))
	return message
}

// decodeLegacy05ForTest mirrors the deployed 0.5 parser: unknown TLVs are
// ignored and the final signature authenticates the public key and link secret.
func decodeLegacy05ForTest(message, password []byte) error {
	if len(message) < 6+ed25519.SignatureSize || !bytes.Equal(message[:4], []byte("meta")) {
		return ErrHandshakeInvalidLength
	}
	bodyLen := int(binary.BigEndian.Uint16(message[4:6]))
	if bodyLen != len(message)-6 {
		return ErrHandshakeInvalidLength
	}
	body := message[6:]
	sig := body[len(body)-ed25519.SignatureSize:]
	fields := body[:len(body)-ed25519.SignatureSize]
	var publicKey ed25519.PublicKey
	var major, minor uint16
	for len(fields) >= 4 {
		op := binary.BigEndian.Uint16(fields[:2])
		fieldLen := int(binary.BigEndian.Uint16(fields[2:4]))
		fields = fields[4:]
		if len(fields) < fieldLen {
			return ErrHandshakeInvalidLength
		}
		field := fields[:fieldLen]
		switch op {
		case metaVersionMajor:
			if len(field) != 2 {
				return ErrHandshakeInvalidLength
			}
			major = binary.BigEndian.Uint16(field)
		case metaVersionMinor:
			if len(field) != 2 {
				return ErrHandshakeInvalidLength
			}
			minor = binary.BigEndian.Uint16(field)
		case metaPublicKey:
			if len(field) != ed25519.PublicKeySize {
				return ErrHandshakeInvalidLength
			}
			publicKey = append(publicKey[:0], field...)
		}
		fields = fields[fieldLen:]
	}
	if len(fields) != 0 || major != ProtocolVersionMajor || minor != ProtocolVersionMinor {
		return ErrHandshakeInvalidLength
	}
	hash, err := legacyHandshakeHash(password, publicKey)
	if err != nil {
		return err
	}
	if !ed25519.Verify(publicKey, hash, sig) {
		return ErrHandshakeInvalidSignature
	}
	return nil
}

func FuzzVersionDecode(f *testing.F) {
	f.Add([]byte("meta\x00\x00"), []byte(""))
	f.Add([]byte("not a handshake"), []byte("secret"))
	f.Fuzz(func(t *testing.T, message, password []byte) {
		// The parser must reject malformed or unauthenticated input without
		// panicking, allocating unbounded memory, or reading beyond the message.
		var decoded version_metadata
		_ = decoded.decode(bytes.NewReader(message), password)
	})
}
