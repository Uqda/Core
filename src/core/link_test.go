package core

import (
	"net/url"
	"testing"
	"time"

	"github.com/yggdrasil-network/yggdrasil-go/src/config"
)

// TestJitteredBackoffDurationBounds characterizes jitteredBackoffDuration's
// contract: the reconnect backoff used by links.add's backoffNow closure.
// Equal jitter should keep the result within [duration/2, duration], and
// duration itself must never exceed max regardless of what was asked for.
func TestJitteredBackoffDurationBounds(t *testing.T) {
	const max = 10 * time.Second

	for _, duration := range []time.Duration{
		0,
		1,
		time.Second,
		2 * time.Second,
		9 * time.Second,
		10 * time.Second,
		time.Hour, // deliberately larger than max
	} {
		for i := 0; i < 200; i++ {
			got := jitteredBackoffDuration(duration, max)
			if got > max {
				t.Fatalf("jitteredBackoffDuration(%s, %s) = %s, exceeds max", duration, max, got)
			}
			capped := duration
			if capped > max {
				capped = max
			}
			half := capped / 2
			if got < half {
				t.Fatalf("jitteredBackoffDuration(%s, %s) = %s, below the guaranteed half (%s)", duration, max, got, half)
			}
			if got > capped {
				t.Fatalf("jitteredBackoffDuration(%s, %s) = %s, exceeds the capped duration (%s)", duration, max, got, capped)
			}
		}
	}
}

// TestIsValidPeerSchemeMatchesDialerFor cross-checks IsValidPeerScheme
// (exported for cmd/yggdrasil's "-checkconf") directly against
// links.dialerFor's own switch statement, using a real Core - dialerFor
// only selects which linkProtocol would handle a scheme, it doesn't
// perform any I/O, so this is safe to call for every candidate without
// actually dialing anything.
func TestIsValidPeerSchemeMatchesDialerFor(t *testing.T) {
	cfg := config.GenerateConfig()
	if err := cfg.GenerateSelfSignedCertificate(); err != nil {
		t.Fatal(err)
	}
	c, err := New(cfg.Certificate, GetLoggerWithPrefix("", false))
	if err != nil {
		t.Fatal(err)
	}
	defer c.Stop()

	for _, scheme := range []string{"tcp", "tls", "socks", "sockstls", "unix", "quic", "ws", "wss", "bogus", ""} {
		u := &url.URL{Scheme: scheme, Host: "localhost:1234"}
		_, dialErr := c.links.dialerFor(u)
		gotRecognised := dialErr == nil
		wantRecognised := IsValidPeerScheme(scheme)
		if gotRecognised != wantRecognised {
			t.Fatalf("scheme %q: dialerFor recognised=%v, IsValidPeerScheme=%v - these must stay in sync", scheme, gotRecognised, wantRecognised)
		}
	}
}

// TestIsValidListenScheme is a literal-values check against
// links.listen's switch statement (see that function and
// IsValidListenScheme's doc comment) - deliberately not cross-verified
// via a live listener the way TestIsValidPeerSchemeMatchesDialerFor is,
// since links.listen actually binds an OS-level socket for every
// recognised scheme rather than just selecting a handler.
func TestIsValidListenScheme(t *testing.T) {
	for scheme, want := range map[string]bool{
		"tcp": true, "tls": true, "unix": true, "quic": true, "ws": true, "wss": true,
		"socks": false, "sockstls": false, "bogus": false, "": false,
	} {
		if got := IsValidListenScheme(scheme); got != want {
			t.Errorf("IsValidListenScheme(%q) = %v, want %v", scheme, got, want)
		}
	}
}

// TestJitteredBackoffDurationVaries is a sanity check that jitter is
// actually happening - not a statistical test, just confirmation that
// repeated calls with the same input don't all return the exact same
// value, which would indicate the randomization was accidentally
// stripped out (e.g. a future edit collapsing this to duration/2+0).
func TestJitteredBackoffDurationVaries(t *testing.T) {
	const duration = 10 * time.Second
	const max = time.Minute

	first := jitteredBackoffDuration(duration, max)
	for i := 0; i < 100; i++ {
		if jitteredBackoffDuration(duration, max) != first {
			return
		}
	}
	t.Fatalf("jitteredBackoffDuration(%s, %s) returned %s on every one of 101 calls - jitter does not appear to be applied", duration, max, first)
}
