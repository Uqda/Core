package identity

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"testing"
	"time"
)

// selfSignedCert mirrors what src/config.GenerateSelfSignedCertificate
// produces closely enough to characterize GenerateTLSConfig's behavior
// without importing the config package (which would be a cyclic-looking
// dependency for a test, even though it wouldn't actually cycle - keeping
// this package's tests self-contained is simpler).
func selfSignedCert(t *testing.T) *tls.Certificate {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "test"},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(time.Hour),
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, pub, priv)
	if err != nil {
		t.Fatal(err)
	}
	return &tls.Certificate{Certificate: [][]byte{der}, PrivateKey: priv}
}

// TestGenerateTLSConfig characterizes the observable behavior this package
// took over from src/core/tls.go: the returned config wraps the given
// certificate, requires TLS 1.3, and - per the design note in tls.go -
// never fails certificate verification at this layer, since that trust
// decision is made afterward at the handshake layer instead.
func TestGenerateTLSConfig(t *testing.T) {
	cert := selfSignedCert(t)
	cfg := GenerateTLSConfig(cert)

	if len(cfg.Certificates) != 1 {
		t.Fatalf("expected exactly one certificate, got %d", len(cfg.Certificates))
	}
	if cfg.Certificates[0].PrivateKey == nil {
		t.Fatal("expected the certificate's private key to be preserved")
	}
	if cfg.MinVersion != tls.VersionTLS13 {
		t.Fatalf("expected MinVersion TLS 1.3, got %x", cfg.MinVersion)
	}
	if !cfg.InsecureSkipVerify {
		t.Fatal("expected InsecureSkipVerify to be true (trust decision happens at the handshake layer, not here)")
	}

	gotCert, err := cfg.GetClientCertificate(nil)
	if err != nil || gotCert != cert {
		t.Fatalf("GetClientCertificate should return the original cert unchanged: got %v, err %v", gotCert, err)
	}
	if err := cfg.VerifyPeerCertificate(nil, nil); err != nil {
		t.Fatalf("VerifyPeerCertificate must always accept (verification happens elsewhere): %v", err)
	}
	if err := cfg.VerifyConnection(tls.ConnectionState{}); err != nil {
		t.Fatalf("VerifyConnection must always accept (verification happens elsewhere): %v", err)
	}
}
