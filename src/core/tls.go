package core

import (
	"crypto/ed25519"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
)

func (c *Core) generateTLSConfig(cert *tls.Certificate) (*tls.Config, error) {
	config := &tls.Config{
		Certificates: []tls.Certificate{*cert},
		ClientAuth:   tls.RequireAnyClientCert,
		GetClientCertificate: func(cri *tls.CertificateRequestInfo) (*tls.Certificate, error) {
			return cert, nil
		},
		VerifyPeerCertificate: c.verifyTLSCertificate,
		VerifyConnection:      c.verifyTLSConnection,
		InsecureSkipVerify:    true,
		MinVersion:            tls.VersionTLS13,
	}
	return config, nil
}

func (c *Core) verifyTLSCertificate(_ [][]byte, _ [][]*x509.Certificate) error {
	return nil
}

func (c *Core) verifyTLSConnection(_ tls.ConnectionState) error {
	return nil
}

// verifyTLSPeerIdentity binds the certificate used by a direct TLS transport
// to the Ed25519 identity authenticated by the UQDA handshake. Certificate
// chains are intentionally not used because UQDA identities are self-issued.
func verifyTLSPeerIdentity(conn net.Conn, expected ed25519.PublicKey) error {
	if tracked, ok := conn.(*linkConn); ok {
		conn = tracked.Conn
	}
	tlsConn, ok := conn.(*tls.Conn)
	if !ok {
		return nil
	}
	state := tlsConn.ConnectionState()
	if len(state.PeerCertificates) != 1 {
		return fmt.Errorf("TLS peer presented %d certificates, want exactly one", len(state.PeerCertificates))
	}
	peerKey, ok := state.PeerCertificates[0].PublicKey.(ed25519.PublicKey)
	if !ok {
		return fmt.Errorf("TLS peer certificate does not contain an Ed25519 key")
	}
	if !peerKey.Equal(expected) {
		return fmt.Errorf("TLS peer certificate does not match UQDA identity")
	}
	return nil
}
