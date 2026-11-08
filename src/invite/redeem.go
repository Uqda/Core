package invite

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"time"
)

// Redeem contacts the inviting node and returns its grant. It tries each
// address in the token and accepts only a server that proves possession of
// the token's public key in the TLS 1.3 handshake.
func Redeem(ctx context.Context, t Token) (Grant, error) {
	key, err := t.publicKey()
	if err != nil {
		return Grant{}, err
	}
	if t.Expires > 0 && time.Now().Unix() > t.Expires {
		return Grant{}, errors.New("this invite has expired; ask for a new one")
	}
	var errs []string
	for _, addr := range t.Addresses {
		g, err := redeemAt(ctx, addr, key, t.Secret)
		if err == nil {
			return g, nil
		}
		var refused *refusedError
		if errors.As(err, &refused) {
			return Grant{}, refused // the right node answered; other addresses will not help
		}
		errs = append(errs, fmt.Sprintf("%s: %v", addr, err))
		if ctx.Err() != nil {
			break
		}
	}
	return Grant{}, fmt.Errorf("could not reach the inviting node (%s). Make sure both devices are online, and that the inviter's firewall allows its invite port", strings.Join(errs, "; "))
}

type refusedError struct{ msg string }

func (e *refusedError) Error() string { return "the inviting node refused: " + e.msg }

func redeemAt(ctx context.Context, addr string, key ed25519.PublicKey, secret string) (Grant, error) {
	dialer := &tls.Dialer{
		NetDialer: &net.Dialer{Timeout: 5 * time.Second},
		Config: &tls.Config{
			MinVersion: tls.VersionTLS13,
			// The node certificate is self-signed; identity comes from the
			// pinned public key instead of a certificate authority.
			InsecureSkipVerify: true, //nolint:gosec // verified by VerifyPeerCertificate below
			VerifyPeerCertificate: func(raw [][]byte, _ [][]*x509.Certificate) error {
				return verifyPinnedKey(raw, key)
			},
		},
	}
	dctx, cancel := context.WithTimeout(ctx, redeemDeadline)
	defer cancel()
	conn, err := dialer.DialContext(dctx, "tcp", addr)
	if err != nil {
		return Grant{}, err
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(redeemDeadline))
	if err := json.NewEncoder(conn).Encode(redeemRequest{Secret: secret}); err != nil {
		return Grant{}, err
	}
	var resp redeemResponse
	if err := json.NewDecoder(io.LimitReader(conn, 64*1024)).Decode(&resp); err != nil {
		return Grant{}, fmt.Errorf("no valid answer: %w", err)
	}
	if !resp.OK || resp.Grant == nil {
		msg := resp.Error
		if msg == "" {
			msg = "invalid or expired invite"
		}
		return Grant{}, &refusedError{msg}
	}
	return *resp.Grant, nil
}

func verifyPinnedKey(raw [][]byte, key ed25519.PublicKey) error {
	if len(raw) == 0 {
		return errors.New("no certificate presented")
	}
	cert, err := x509.ParseCertificate(raw[0])
	if err != nil {
		return err
	}
	pub, ok := cert.PublicKey.(ed25519.PublicKey)
	if !ok || !bytes.Equal(pub, key) {
		return errors.New("the node at this address is not the one that created the invite")
	}
	// TLS 1.3 CertificateVerify, already checked by crypto/tls against this
	// certificate's key, proves the server holds the matching private key.
	return nil
}
