package core

// Connection metadata implements the Yggdrasil wire protocol. Product release
// identity is defined separately in src/version.

import (
	"bytes"
	"crypto/ed25519"
	"encoding/binary"
	"io"

	"golang.org/x/crypto/blake2b"
)

// This is the version-specific metadata exchanged at the start of a connection.
// It begins with the four-byte "meta" preamble and a uint16 payload length.
// Length-delimited fields carry protocol versions, the public key and priority.
type versionMetadata struct {
	majorVer  uint16
	minorVer  uint16
	publicKey ed25519.PublicKey
	priority  uint8
}

const (
	// ProtocolVersionMajor identifies the major Yggdrasil wire protocol version.
	ProtocolVersionMajor uint16 = 0
	// ProtocolVersionMinor identifies the minor Yggdrasil wire protocol version.
	ProtocolVersionMinor uint16 = 5
)

// Once a major/minor version is released, it is not safe to change any of these
// (including their ordering), it is only safe to add new ones.
const (
	metaVersionMajor uint16 = iota // uint16
	metaVersionMinor               // uint16
	metaPublicKey                  // [32]byte
	metaPriority                   // uint8
)

type handshakeError string

func (e handshakeError) Error() string { return string(e) }

// Handshake errors describe rejected peer metadata.
const (
	ErrHandshakeInvalidPreamble   = handshakeError("invalid handshake: remote peer did not send a Yggdrasil protocol preamble")
	ErrHandshakeInvalidLength     = handshakeError("invalid handshake length, possible version mismatch")
	ErrHandshakeInvalidPassword   = handshakeError("invalid password supplied, check your config")
	ErrHandshakeHashFailure       = handshakeError("invalid hash length")
	ErrHandshakeIncorrectPassword = handshakeError("password does not match remote side")
)

// versionGetBaseMetadata returns metadata with the supported protocol version.
func versionGetBaseMetadata() versionMetadata {
	return versionMetadata{
		majorVer: ProtocolVersionMajor,
		minorVer: ProtocolVersionMinor,
	}
}

// encode serializes version metadata in its wire format.
func (m *versionMetadata) encode(privateKey ed25519.PrivateKey, password []byte) ([]byte, error) {
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

	hasher, err := blake2b.New512(password)
	if err != nil {
		return nil, err
	}
	n, err := hasher.Write(m.publicKey)
	if err != nil {
		return nil, err
	}
	if n != ed25519.PublicKeySize {
		return nil, ErrHandshakeHashFailure
	}
	hash := hasher.Sum(nil)
	bs = append(bs, ed25519.Sign(privateKey, hash)...)

	binary.BigEndian.PutUint16(bs[4:6], uint16(len(bs)-6))
	return bs, nil
}

// decode parses version metadata from its wire format.
func (m *versionMetadata) decode(r io.Reader, password []byte) error {
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
	sig := bs[len(bs)-ed25519.SignatureSize:]
	bs = bs[:len(bs)-ed25519.SignatureSize]

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
		}
		bs = bs[oplen:]
	}
	if len(bs) != 0 {
		return ErrHandshakeInvalidLength
	}

	hasher, err := blake2b.New512(password)
	if err != nil {
		return ErrHandshakeInvalidPassword
	}
	n, err := hasher.Write(m.publicKey)
	if err != nil || n != ed25519.PublicKeySize {
		return ErrHandshakeHashFailure
	}
	hash := hasher.Sum(nil)
	if !ed25519.Verify(m.publicKey, hash, sig) {
		return ErrHandshakeIncorrectPassword
	}
	return nil
}

// check reports whether the peer metadata is compatible and complete.
func (m *versionMetadata) check() bool {
	switch {
	case m.majorVer != ProtocolVersionMajor:
		return false
	case m.minorVer != ProtocolVersionMinor:
		return false
	case len(m.publicKey) != ed25519.PublicKeySize:
		return false
	default:
		return true
	}
}
