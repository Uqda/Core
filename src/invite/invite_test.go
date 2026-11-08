package invite

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"errors"
	"net"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Uqda/Core/src/config"
)

type testLogger struct{}

func (testLogger) Infof(string, ...interface{})  {}
func (testLogger) Warnf(string, ...interface{})  {}
func (testLogger) Debugf(string, ...interface{}) {}

func newManager(t *testing.T, grant Grant) (*Manager, *config.NodeConfig) {
	t.Helper()
	cfg := config.GenerateConfig()
	if err := cfg.GenerateSelfSignedCertificate(); err != nil {
		t.Fatal(err)
	}
	pub := ed25519.PrivateKey(cfg.PrivateKey).Public().(ed25519.PublicKey)
	m := New(cfg.Certificate, pub, 0, func(net.IP) Grant { return grant }, testLogger{})
	m.addrs = func(port int) []string { return []string{net.JoinHostPort("127.0.0.1", strconv.Itoa(port))} }
	t.Cleanup(m.Stop)
	return m, cfg
}

func mustCreate(t *testing.T, m *Manager, ttl time.Duration, uses int) Token {
	t.Helper()
	c, err := m.Create(ttl, uses, "home")
	if err != nil {
		t.Fatal(err)
	}
	tok, err := Decode(c.Token)
	if err != nil {
		t.Fatal(err)
	}
	return tok
}

func redeem(t Token) (Grant, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return Redeem(ctx, t)
}

func TestTokenRoundTripAndValidation(t *testing.T) {
	tok := Token{Key: strings.Repeat("ab", 32), Addresses: []string{"192.0.2.1:1234"}, Secret: strings.Repeat("A", 43), Expires: 1}
	got, err := Decode(tok.Encode())
	if err != nil || got.Key != tok.Key || got.Addresses[0] != tok.Addresses[0] {
		t.Fatalf("round trip failed: %+v %v", got, err)
	}
	bad := []string{
		"",
		"hello",
		TokenPrefix + "!!!",
		Token{Key: "zz", Addresses: tok.Addresses, Secret: tok.Secret}.Encode(),
		Token{Key: tok.Key, Addresses: nil, Secret: tok.Secret}.Encode(),
		Token{Key: tok.Key, Addresses: []string{"no-port"}, Secret: tok.Secret}.Encode(),
		Token{Key: tok.Key, Addresses: tok.Addresses, Secret: "short"}.Encode(),
		TokenPrefix + strings.Repeat("A", maxTokenSize),
	}
	for _, s := range bad {
		if _, err := Decode(s); err == nil {
			t.Fatalf("accepted invalid invite %q", s)
		}
	}
}

func TestInviteRedeemsOnceAndCarriesNoSecretsInTheToken(t *testing.T) {
	grant := Grant{GroupPassword: "correct horse battery staple", Peers: []string{"tls://192.0.2.10:443"}}
	m, _ := newManager(t, grant)
	c, err := m.Create(time.Minute, 1, "home")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(c.Token, "horse") || strings.Contains(c.Token, "192.0.2.10") {
		t.Fatal("the invite token leaks the grant")
	}
	tok, _ := Decode(c.Token)

	got, err := redeem(tok)
	if err != nil {
		t.Fatal(err)
	}
	if got.GroupPassword != grant.GroupPassword || len(got.Peers) != 1 || got.Name != "home" {
		t.Fatalf("unexpected grant: %+v", got)
	}
	if _, err := redeem(tok); err == nil {
		t.Fatal("a single-use invite was redeemed twice")
	}
	if m.Active() != 0 {
		t.Fatal("used invite is still active")
	}
	m.mu.Lock()
	open := m.listener != nil
	m.mu.Unlock()
	if open {
		t.Fatal("redemption port stayed open with no active invites")
	}
}

func TestMultiUseInvite(t *testing.T) {
	m, _ := newManager(t, Grant{})
	tok := mustCreate(t, m, time.Minute, 3)
	for i := 0; i < 3; i++ {
		if _, err := redeem(tok); err != nil {
			t.Fatalf("use %d: %v", i+1, err)
		}
	}
	if _, err := redeem(tok); err == nil {
		t.Fatal("invite worked beyond its use count")
	}
}

func TestExpiredInviteIsRefused(t *testing.T) {
	m, _ := newManager(t, Grant{})
	var offset atomic.Int64 // read by the server goroutine, so keep it atomic
	m.mu.Lock()
	m.now = func() time.Time { return time.Now().Add(time.Duration(offset.Load())) }
	m.mu.Unlock()
	tok := mustCreate(t, m, time.Minute, 1)
	// Skip the client-side expiry check: the server must refuse on its own.
	tok.Expires = 0
	// Hold the listener open by keeping a second, longer invite active.
	if _, err := m.Create(time.Hour, 1, ""); err != nil {
		t.Fatal(err)
	}
	offset.Store(int64(2 * time.Minute))
	_, err := redeem(tok)
	var refused *refusedError
	if !errors.As(err, &refused) {
		t.Fatalf("expired invite was not refused by the server: %v", err)
	}
	if m.Active() != 1 {
		t.Fatalf("expected only the unexpired invite to remain, got %d", m.Active())
	}
}

func TestWrongSecretFailsAndIsRateLimited(t *testing.T) {
	m, _ := newManager(t, Grant{GroupPassword: "secret"})
	tok := mustCreate(t, m, time.Minute, 1)
	good := tok.Secret
	tok.Secret = strings.Repeat("B", 43)
	for i := 0; i < maxFailuresPerIP; i++ {
		_, err := redeem(tok)
		var refused *refusedError
		if !errors.As(err, &refused) {
			t.Fatalf("attempt %d: expected a refusal, got %v", i+1, err)
		}
	}
	tok.Secret = good
	_, err := redeem(tok)
	if err == nil || !strings.Contains(err.Error(), "too many failed attempts") {
		t.Fatalf("correct secret after the limit should be blocked, got %v", err)
	}
	if m.Active() != 1 {
		t.Fatal("blocked attempts must not consume the invite")
	}
}

func TestImpostorNodeIsRejected(t *testing.T) {
	// An attacker controls the address in the invite but not the key.
	real, _ := newManager(t, Grant{GroupPassword: "secret"})
	impostor, _ := newManager(t, Grant{GroupPassword: "stolen"})
	tok := mustCreate(t, real, time.Minute, 1)
	fake := mustCreate(t, impostor, time.Minute, 1)
	tok.Addresses = fake.Addresses // point the real invite at the impostor
	_, err := redeem(tok)
	if err == nil || !strings.Contains(err.Error(), "not the one that created the invite") {
		t.Fatalf("impostor was accepted or failed for the wrong reason: %v", err)
	}
	if real.Active() != 1 {
		t.Fatal("the real invite should still be usable")
	}
}

func TestCancelAllRevokes(t *testing.T) {
	m, _ := newManager(t, Grant{})
	tok := mustCreate(t, m, time.Minute, 1)
	if n := m.CancelAll(); n != 1 {
		t.Fatalf("cancelled %d", n)
	}
	if _, err := redeem(tok); err == nil {
		t.Fatal("cancelled invite still works")
	}
}

func TestCreateLimits(t *testing.T) {
	m, _ := newManager(t, Grant{})
	if _, err := m.Create(MaxTTL+time.Second, 1, ""); err == nil {
		t.Fatal("accepted an expiry beyond the maximum")
	}
	if _, err := m.Create(time.Minute, MaxUses+1, ""); err == nil {
		t.Fatal("accepted too many uses")
	}
	for i := 0; i < maxActiveInvites; i++ {
		if _, err := m.Create(time.Minute, 1, ""); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := m.Create(time.Minute, 1, ""); err == nil {
		t.Fatal("accepted more than the maximum number of active invites")
	}
}

func TestApplyKeepsIdentityAndMergesPeers(t *testing.T) {
	cfg := config.GenerateConfig()
	key := hex.EncodeToString(cfg.PrivateKey)
	cfg.Peers = []string{"tls://a:1"}
	changes := Apply(cfg, Grant{GroupPassword: "g", Peers: []string{"tls://a:1", "tls://b:2"}})
	if hex.EncodeToString(cfg.PrivateKey) != key {
		t.Fatal("identity changed")
	}
	if cfg.GroupPassword != "g" || len(cfg.Peers) != 2 || len(changes) != 2 {
		t.Fatalf("unexpected result: %+v %v", cfg.Peers, changes)
	}
	if again := Apply(cfg, Grant{GroupPassword: "g", Peers: []string{"tls://b:2"}}); len(again) != 0 {
		t.Fatalf("re-applying the same grant reported changes: %v", again)
	}
}
