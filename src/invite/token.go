// Package invite lets one node invite another into its network without
// sharing secrets in the invitation itself.
//
// An invite token carries the inviting node's public key, the addresses
// where it is listening for redemption, and a random one-time secret. The
// joining node connects over TLS 1.3, checks that the server proves
// possession of that public key, and presents the secret. Only then does the
// inviting node send its group password and peers, inside the encrypted
// connection. Tokens expire, are single-use by default, and stop working as
// soon as the inviting node restarts, because only a hash of the secret is
// kept in memory.
package invite

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strings"
)

// TokenPrefix starts every version 1 invite token.
const TokenPrefix = "uqda-invite-1:"

const (
	secretSize   = 32
	maxAddresses = 8
	maxTokenSize = 2048
	maxNameSize  = 64
)

// Token is a decoded invitation.
type Token struct {
	Key       string   `json:"k"`           // inviting node's public key, hex
	Addresses []string `json:"a"`           // host:port candidates for redemption
	Secret    string   `json:"s"`           // base64url one-time secret
	Expires   int64    `json:"e"`           // Unix seconds; enforced by the inviter
	Name      string   `json:"n,omitempty"` // optional label chosen by the inviter
}

// Encode returns the token as a single copyable string.
func (t Token) Encode() string {
	b, _ := json.Marshal(t)
	return TokenPrefix + base64.RawURLEncoding.EncodeToString(b)
}

// Decode parses and validates an invite token string.
func Decode(s string) (Token, error) {
	var t Token
	s = strings.TrimSpace(s)
	if len(s) > maxTokenSize {
		return t, errors.New("invite is too long")
	}
	if !strings.HasPrefix(s, TokenPrefix) {
		return t, errors.New("this is not an Uqda invite (it should start with " + TokenPrefix + ")")
	}
	raw, err := base64.RawURLEncoding.DecodeString(strings.TrimPrefix(s, TokenPrefix))
	if err != nil {
		return t, errors.New("invite is damaged; copy the whole invite again")
	}
	if err := json.Unmarshal(raw, &t); err != nil {
		return t, errors.New("invite is damaged; copy the whole invite again")
	}
	if _, err := t.publicKey(); err != nil {
		return t, err
	}
	if _, err := t.secret(); err != nil {
		return t, err
	}
	if len(t.Addresses) == 0 || len(t.Addresses) > maxAddresses {
		return t, fmt.Errorf("invite must list 1 to %d addresses", maxAddresses)
	}
	for _, a := range t.Addresses {
		if _, _, err := net.SplitHostPort(a); err != nil {
			return t, fmt.Errorf("invite address %q is invalid", a)
		}
	}
	if len(t.Name) > maxNameSize {
		return t, errors.New("invite name is too long")
	}
	return t, nil
}

func (t Token) publicKey() (ed25519.PublicKey, error) {
	k, err := hex.DecodeString(t.Key)
	if err != nil || len(k) != ed25519.PublicKeySize {
		return nil, errors.New("invite has an invalid node key")
	}
	return ed25519.PublicKey(k), nil
}

func (t Token) secret() ([]byte, error) {
	s, err := base64.RawURLEncoding.DecodeString(t.Secret)
	if err != nil || len(s) != secretSize {
		return nil, errors.New("invite has an invalid secret")
	}
	return s, nil
}

// Grant is what a successful redemption returns to the joining node.
type Grant struct {
	GroupPassword  string   `json:"group_password"`
	Peers          []string `json:"peers"`
	InviterKey     string   `json:"inviter_key"`
	InviterAddress string   `json:"inviter_address"`
	Name           string   `json:"name,omitempty"`
}

type redeemRequest struct {
	Secret string `json:"secret"`
}

type redeemResponse struct {
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
	Grant *Grant `json:"grant,omitempty"`
}
