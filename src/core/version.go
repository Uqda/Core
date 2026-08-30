package core

// This file contains the version metadata struct
// Used in the initial connection setup and key exchange
// Some of this could arguably go in wire.go instead

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/binary"
	"io"

	"golang.org/x/crypto/blake2b"
)

// This is the version-specific metadata exchanged at the start of a connection.
// It must always begin with the 4 bytes "meta" and a wire formatted uint64 major version number.
// The current version also includes a minor version number, and the box/sig/link keys that need to be exchanged to open a connection.
type version_metadata struct {
	majorVer  uint16
	minorVer  uint16
	publicKey ed25519.PublicKey
	priority  uint8
	nonce     [32]byte
	wire      []byte
}

const (
	ProtocolVersionMajor uint16 = 0
	ProtocolVersionMinor uint16 = 6
)

// Once a major/minor version is released, it is not safe to change any of these
// (including their ordering), it is only safe to add new ones.
const (
	metaVersionMajor uint16 = iota // uint16
	metaVersionMinor               // uint16
	metaPublicKey                  // [32]byte
	metaPriority                   // uint8
	metaNonce                      // [32]byte
)

type handshakeError string

func (e handshakeError) Error() string { return string(e) }

const ErrHandshakeInvalidPreamble = handshakeError("invalid handshake, remote side is not UQDA")
const ErrHandshakeInvalidLength = handshakeError("invalid handshake length, possible version mismatch")
const ErrHandshakeInvalidPassword = handshakeError("invalid password supplied, check your config")
const ErrHandshakeHashFailure = handshakeError("invalid hash length")
const ErrHandshakeIncorrectPassword = handshakeError("password does not match remote side")
const ErrHandshakeInvalidSignature = handshakeError("handshake signature verification failed")
const ErrHandshakeInvalidConfirmation = handshakeError("handshake confirmation verification failed")

const confirmationSize = 4 + ed25519.SignatureSize

var handshakeDomain = []byte("UQDA-HANDSHAKE-V1\x00")
var confirmationDomain = []byte("UQDA-CONFIRM-V1\x00")

// Gets a base metadata with no keys set, but with the correct version numbers.
func version_getBaseMetadata() version_metadata {
	return version_metadata{
		majorVer: ProtocolVersionMajor,
		minorVer: ProtocolVersionMinor,
	}
}

// Encodes version metadata into its wire format.
func (m *version_metadata) encode(privateKey ed25519.PrivateKey, password []byte) ([]byte, error) {
	if _, err := rand.Read(m.nonce[:]); err != nil {
		return nil, err
	}
	bs := make([]byte, 0, 64)
	bs = append(bs, 'm', 'e', 't', 'a')
	bs = append(bs, 0, 0) // Remaining message length

	bs = binary.BigEndian.AppendUint16(bs, metaVersionMajor)
	bs = binary.BigEndian.AppendUint16(bs, 2)
	bs = binary.BigEndian.AppendUint16(bs, m.majorVer)

	bs = binary.BigEndian.AppendUint16(bs, metaVersionMinor)
	bs = binary.BigEndian.AppendUint16(bs, 2)
	bs = binary.BigEndian.AppendUint16(bs, m.minorVer)

	bs = binary.BigEndian.AppendUint16(bs, metaPublicKey)
	bs = binary.BigEndian.AppendUint16(bs, ed25519.PublicKeySize)
	bs = append(bs, m.publicKey[:]...)

	bs = binary.BigEndian.AppendUint16(bs, metaPriority)
	bs = binary.BigEndian.AppendUint16(bs, 1)
	bs = append(bs, m.priority)

	bs = binary.BigEndian.AppendUint16(bs, metaNonce)
	bs = binary.BigEndian.AppendUint16(bs, uint16(len(m.nonce)))
	bs = append(bs, m.nonce[:]...)

	hash, err := handshakeHash(password, bs[6:])
	if err != nil {
		return nil, err
	}
	bs = append(bs, ed25519.Sign(privateKey, hash)...)

	binary.BigEndian.PutUint16(bs[4:6], uint16(len(bs)-6))
	m.wire = append(m.wire[:0], bs...)
	return bs, nil
}

// Decodes version metadata from its wire format into the struct.
func (m *version_metadata) decode(r io.Reader, password []byte) error {
	bh := [6]byte{}
	if _, err := io.ReadFull(r, bh[:]); err != nil {
		return err
	}
	meta := [4]byte{'m', 'e', 't', 'a'}
	if !bytes.Equal(bh[:4], meta[:]) {
		return ErrHandshakeInvalidPreamble
	}
	hl := binary.BigEndian.Uint16(bh[4:6])
	if hl < ed25519.SignatureSize {
		return ErrHandshakeInvalidLength
	}
	bs := make([]byte, hl)
	if _, err := io.ReadFull(r, bs); err != nil {
		return err
	}
	m.wire = append(m.wire[:0], bh[:]...)
	m.wire = append(m.wire, bs...)
	sig := bs[len(bs)-ed25519.SignatureSize:]
	bs = bs[:len(bs)-ed25519.SignatureSize]
	signedFields := append([]byte(nil), bs...)

	for len(bs) >= 4 {
		op := binary.BigEndian.Uint16(bs[:2])
		oplen := int(binary.BigEndian.Uint16(bs[2:4]))
		if bs = bs[4:]; len(bs) < oplen {
			return ErrHandshakeInvalidLength
		}
		field := bs[:oplen]
		switch op {
		case metaVersionMajor:
			if len(field) != 2 {
				return ErrHandshakeInvalidLength
			}
			m.majorVer = binary.BigEndian.Uint16(field)

		case metaVersionMinor:
			if len(field) != 2 {
				return ErrHandshakeInvalidLength
			}
			m.minorVer = binary.BigEndian.Uint16(field)

		case metaPublicKey:
			if len(field) != ed25519.PublicKeySize {
				return ErrHandshakeInvalidLength
			}
			m.publicKey = append(m.publicKey[:0], field...)

		case metaPriority:
			if len(field) != 1 {
				return ErrHandshakeInvalidLength
			}
			m.priority = field[0]

		case metaNonce:
			if len(field) != len(m.nonce) {
				return ErrHandshakeInvalidLength
			}
			copy(m.nonce[:], field)
		}
		bs = bs[oplen:]
	}
	if len(bs) != 0 {
		return ErrHandshakeInvalidLength
	}

	hash, err := handshakeHash(password, signedFields)
	if err != nil {
		return ErrHandshakeInvalidPassword
	}
	if !ed25519.Verify(m.publicKey, hash, sig) {
		return ErrHandshakeInvalidSignature
	}
	return nil
}

func handshakeHash(password, payload []byte) ([]byte, error) {
	hasher, err := blake2b.New512(password)
	if err != nil {
		return nil, err
	}
	if _, err = hasher.Write(handshakeDomain); err != nil {
		return nil, err
	}
	if _, err = hasher.Write(payload); err != nil {
		return nil, err
	}
	return hasher.Sum(nil), nil
}

func confirmationHash(password []byte, local, remote *version_metadata) ([]byte, error) {
	if len(local.wire) == 0 || len(remote.wire) == 0 {
		return nil, ErrHandshakeInvalidConfirmation
	}
	hasher, err := blake2b.New512(password)
	if err != nil {
		return nil, err
	}
	if _, err = hasher.Write(confirmationDomain); err != nil {
		return nil, err
	}
	first, second := local, remote
	if bytes.Compare(first.publicKey, second.publicKey) > 0 {
		first, second = second, first
	}
	if _, err = hasher.Write(first.wire); err != nil {
		return nil, err
	}
	if _, err = hasher.Write(second.wire); err != nil {
		return nil, err
	}
	return hasher.Sum(nil), nil
}

func encodeConfirmation(privateKey ed25519.PrivateKey, password []byte, local, remote *version_metadata) ([]byte, error) {
	hash, err := confirmationHash(password, local, remote)
	if err != nil {
		return nil, err
	}
	bs := make([]byte, 0, confirmationSize)
	bs = append(bs, 'c', 'o', 'n', 'f')
	bs = append(bs, ed25519.Sign(privateKey, hash)...)
	return bs, nil
}

func decodeConfirmation(r io.Reader, publicKey ed25519.PublicKey, password []byte, local, remote *version_metadata) error {
	bs := make([]byte, confirmationSize)
	if _, err := io.ReadFull(r, bs); err != nil {
		return err
	}
	if !bytes.Equal(bs[:4], []byte{'c', 'o', 'n', 'f'}) {
		return ErrHandshakeInvalidConfirmation
	}
	hash, err := confirmationHash(password, local, remote)
	if err != nil {
		return err
	}
	if !ed25519.Verify(publicKey, hash, bs[4:]) {
		return ErrHandshakeInvalidConfirmation
	}
	return nil
}

// Checks that the "meta" bytes and the version numbers are the expected values.
func (m *version_metadata) check() bool {
	switch {
	case m.majorVer != ProtocolVersionMajor:
		return false
	case m.minorVer != ProtocolVersionMinor:
		return false
	case len(m.publicKey) != ed25519.PublicKeySize:
		return false
	case bytes.Equal(m.nonce[:], make([]byte, len(m.nonce))):
		return false
	default:
		return true
	}
}
