package core

import (
	"crypto/ed25519"
	"errors"
	"fmt"
	"io"
	"net"
)

// ErrHandshakeNotFresh rejects a peer that did not take part in the Uqda
// fresh handshake while RequireFreshHandshake is enabled.
var ErrHandshakeNotFresh = errors.New("peer did not prove its key for this connection (fresh handshake required); it may be a Yggdrasil or older Uqda node")

// ErrHandshakeStaleProof means the peer's fresh-handshake signature did not
// verify: a replayed handshake or a peer without the private key.
var ErrHandshakeStaleProof = errors.New("peer failed to prove its key for this connection (possible replayed handshake)")

const freshContext = "uqda fresh handshake v1"

// freshMessage is what a node signs to prove it saw the peer's nonce on
// this connection. The signer passes the peer's values first.
func freshMessage(peerNonce, ownNonce []byte, peerKey, ownKey ed25519.PublicKey) []byte {
	m := make([]byte, 0, len(freshContext)+2*freshNonceSize+2*ed25519.PublicKeySize)
	m = append(m, freshContext...)
	m = append(m, peerNonce...)
	m = append(m, ownNonce...)
	m = append(m, peerKey...)
	m = append(m, ownKey...)
	return m
}

// freshExchange sends our signature over the peer's nonce and checks the
// peer's signature over ours. Both sides write before reading; each message
// is a single 64-byte signature, well under any socket buffer.
func freshExchange(conn net.Conn, secret ed25519.PrivateKey, own, peer ed25519.PublicKey, ownNonce, peerNonce []byte) error {
	if len(ownNonce) != freshNonceSize || len(peerNonce) != freshNonceSize {
		return ErrHandshakeStaleProof
	}
	sig := ed25519.Sign(secret, freshMessage(peerNonce, ownNonce, peer, own))
	if _, err := conn.Write(sig); err != nil {
		return fmt.Errorf("write fresh handshake: %w", err)
	}
	theirs := make([]byte, ed25519.SignatureSize)
	if _, err := io.ReadFull(conn, theirs); err != nil {
		return fmt.Errorf("read fresh handshake: %w", err)
	}
	// The peer signed (our nonce, its nonce, our key, its key).
	if !ed25519.Verify(peer, freshMessage(ownNonce, peerNonce, own, peer), theirs) {
		return ErrHandshakeStaleProof
	}
	return nil
}
