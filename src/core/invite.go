package core

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

const inviteTokenPrefix = "uqda-invite-v1-"

// InviteToken is the JSON payload embedded in a shareable invite string.
type InviteToken struct {
	V        int      `json:"v"`
	Net      string   `json:"net"`
	Peers    []string `json:"peers"`
	Password string   `json:"password"`
	OwnerKey string   `json:"owner_key"`
	Expires  int64    `json:"expires"`
	// Admin is optional: owner's admin socket (e.g. tcp://host:9001) for inviteRegister.
	Admin string `json:"admin,omitempty"`
}

// EncodeInviteToken serializes the token with the standard prefix.
func EncodeInviteToken(t *InviteToken) (string, error) {
	if t == nil {
		return "", errors.New("nil invite token")
	}
	b, err := json.Marshal(t)
	if err != nil {
		return "", err
	}
	return inviteTokenPrefix + base64.RawURLEncoding.EncodeToString(b), nil
}

// DecodeInviteToken parses an invite string (with or without prefix).
func DecodeInviteToken(s string) (*InviteToken, error) {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, inviteTokenPrefix)
	raw, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		// Accept standard base64 with padding
		raw, err = base64.StdEncoding.DecodeString(s)
		if err != nil {
			return nil, fmt.Errorf("invite token: invalid base64: %w", err)
		}
	}
	var t InviteToken
	if err := json.Unmarshal(raw, &t); err != nil {
		return nil, fmt.Errorf("invite token: invalid JSON: %w", err)
	}
	return &t, nil
}

// ValidateInviteToken checks version, required fields, and expiry.
func ValidateInviteToken(t *InviteToken) error {
	if t == nil {
		return errors.New("nil invite token")
	}
	if t.V != 1 {
		return fmt.Errorf("unsupported invite version %d", t.V)
	}
	if strings.TrimSpace(t.Net) == "" {
		return errors.New("invite token: missing net name")
	}
	if len(t.Peers) == 0 {
		return errors.New("invite token: no peers")
	}
	if len(t.Password) < 16 {
		return errors.New("invite token: invalid password")
	}
	if _, err := decodeHexPubKey(t.OwnerKey); err != nil {
		return fmt.Errorf("invite token owner_key: %w", err)
	}
	if t.Expires > 0 && time.Unix(t.Expires, 0).Before(time.Now()) {
		return errors.New("invite token has expired")
	}
	return nil
}

// ValidateInviteOwner checks that the token was issued for this node's public key.
func ValidateInviteOwner(t *InviteToken, owner ed25519.PublicKey) error {
	if err := ValidateInviteToken(t); err != nil {
		return err
	}
	want, err := decodeHexPubKey(t.OwnerKey)
	if err != nil {
		return err
	}
	if !bytesEqualPub(owner, want) {
		return errors.New("invite token owner_key does not match this node")
	}
	return nil
}

func decodeHexPubKey(h string) (ed25519.PublicKey, error) {
	raw, err := hex.DecodeString(strings.TrimSpace(h))
	if err != nil {
		return nil, err
	}
	if len(raw) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("public key must be %d bytes", ed25519.PublicKeySize)
	}
	return ed25519.PublicKey(raw), nil
}

func bytesEqualPub(a, b ed25519.PublicKey) bool {
	if len(a) != len(b) {
		return false
	}
	var v byte
	for i := range a {
		v |= a[i] ^ b[i]
	}
	return v == 0
}
