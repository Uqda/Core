// Package identity holds node-identity-related helpers that don't belong
// to core's connection/routing orchestration: today, generating the TLS
// configuration peer transports (TCP+TLS, QUIC, WSS) wrap around a node's
// self-signed certificate. See ../../RESTRUCTURING.md for the target
// package layout this is the first slice of.
package identity

import (
	"crypto/tls"
	"crypto/x509"
)

// GenerateTLSConfig builds the tls.Config used to wrap peer connections
// for transports that use TLS (TCP+TLS, QUIC, WSS).
//
// Peer authentication does not happen here: it happens at the handshake
// layer, via the signed "meta" exchange over each node's ed25519 identity
// key (see src/core/version.go), which runs immediately after the
// transport connects regardless of which transport was used. This TLS
// config intentionally skips X.509 certificate-chain verification
// (InsecureSkipVerify, and VerifyPeerCertificate/VerifyConnection callbacks
// that unconditionally return nil) because TLS here provides transport
// encryption over a self-signed, per-node certificate, not the trust
// decision itself - that's a deliberate, pre-existing design choice being
// preserved as-is, not a defect introduced by this package.
func GenerateTLSConfig(cert *tls.Certificate) *tls.Config {
	return &tls.Config{
		Certificates: []tls.Certificate{*cert},
		ClientAuth:   tls.NoClientCert,
		GetClientCertificate: func(*tls.CertificateRequestInfo) (*tls.Certificate, error) {
			return cert, nil
		},
		VerifyPeerCertificate: func([][]byte, [][]*x509.Certificate) error { return nil },
		VerifyConnection:      func(tls.ConnectionState) error { return nil },
		InsecureSkipVerify:    true,
		MinVersion:            tls.VersionTLS13,
	}
}
