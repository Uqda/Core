package admin

import (
	"bytes"
	"net"
	"strings"
	"testing"

	"github.com/gologme/log"
)

func newTestLogger() (*log.Logger, *bytes.Buffer) {
	var buf bytes.Buffer
	l := log.New(&buf, "", 0)
	l.EnableLevel("warn")
	return l, &buf
}

func TestWarnIfListeningPubliclyLoopback(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	logger, buf := newTestLogger()
	a := &AdminSocket{log: logger, listener: ln}
	a.warnIfListeningPublicly()

	if buf.Len() != 0 {
		t.Fatalf("expected no warning for a loopback listener, got: %s", buf.String())
	}
}

func TestWarnIfListeningPubliclyUnspecified(t *testing.T) {
	// Binding to 0.0.0.0 (the "unspecified" address) is exactly the
	// non-loopback case an operator hits by reconfiguring AdminListen
	// without thinking about exposure - it's still safe to bind in a test
	// (nothing outside this machine can reach it purely by us binding to
	// this address locally), but warnIfListeningPublicly must still flag it.
	ln, err := net.Listen("tcp", "0.0.0.0:0")
	if err != nil {
		t.Skipf("could not bind 0.0.0.0: %v", err)
	}
	defer ln.Close()

	logger, buf := newTestLogger()
	a := &AdminSocket{log: logger, listener: ln}
	a.warnIfListeningPublicly()

	out := buf.String()
	if !strings.Contains(out, "WARNING") {
		t.Fatalf("expected a warning for a non-loopback listener, got: %q", out)
	}
	if !strings.Contains(strings.ToLower(out), "authentication") {
		t.Fatalf("expected the warning to explain the lack of authentication, got: %q", out)
	}
}

func TestWarnIfListeningPubliclyUnixSocket(t *testing.T) {
	// Non-TCP listeners (unix sockets) are out of scope for this check -
	// their protection is the socket file's own permissions, not address
	// binding - so this must be a silent no-op rather than a type-assertion
	// panic.
	logger, buf := newTestLogger()
	a := &AdminSocket{log: logger, listener: &fakeUnixLikeListener{}}
	a.warnIfListeningPublicly()

	if buf.Len() != 0 {
		t.Fatalf("expected no warning for a non-TCP listener, got: %s", buf.String())
	}
}

// fakeUnixLikeListener is a minimal net.Listener whose Addr() is not a
// *net.TCPAddr, standing in for a unix socket listener without needing a
// real filesystem path (which isn't available identically on every CI
// platform this might run on).
type fakeUnixLikeListener struct{}

func (f *fakeUnixLikeListener) Accept() (net.Conn, error) { panic("not implemented") }
func (f *fakeUnixLikeListener) Close() error              { return nil }
func (f *fakeUnixLikeListener) Addr() net.Addr            { return fakeAddr{} }

type fakeAddr struct{}

func (fakeAddr) Network() string { return "unix" }
func (fakeAddr) String() string  { return "/run/uqda/admin.sock" }
