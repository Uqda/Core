package core

import (
	"errors"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func writeSecret(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "peer.password")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLinkPasswordSources(t *testing.T) {
	file := writeSecret(t, "from-file\n")
	cases := []struct {
		query string
		want  string
		bad   bool
	}{
		{"", "", false},
		{"password=inline", "inline", false},
		{"password_file=" + url.QueryEscape(file), "from-file", false},
		{"password=a&password_file=" + url.QueryEscape(file), "", true},
		{"password_file=" + url.QueryEscape(filepath.Join(t.TempDir(), "missing")), "", true},
		{"password_file=" + url.QueryEscape(writeSecret(t, "\n")), "", true},
		{"password=" + strings.Repeat("x", 65), "", true},
		{"password_file=" + url.QueryEscape(writeSecret(t, strings.Repeat("y", 65))), "", true},
	}
	for _, tc := range cases {
		q, _ := url.ParseQuery(tc.query)
		got, err := linkPassword(q)
		if tc.bad {
			if err == nil || !errors.Is(err, ErrLinkPasswordInvalid) {
				t.Fatalf("%q: expected ErrLinkPasswordInvalid, got %v", tc.query, err)
			}
			continue
		}
		if err != nil || string(got) != tc.want {
			t.Fatalf("%q: got %q, %v; want %q", tc.query, got, err, tc.want)
		}
	}
}

// A listener and a peer that read the same password from files peer up.
func TestPeeringWithPasswordFiles(t *testing.T) {
	a, b := newCoreWith(t), newCoreWith(t)
	secret := writeSecret(t, "shared-link-secret\n")
	l, err := a.Listen(mustParseURL(t, "tcp://127.0.0.1:0?password_file="+url.QueryEscape(secret)), "")
	requireNoError(t, err)
	requireNoError(t, b.CallPeer(mustParseURL(t, "tcp://"+l.Addr().String()+"?password_file="+url.QueryEscape(secret)), ""))
	if !waitForCondition(5*time.Second, func() bool { _, ok := hasPeer(a, b.PublicKey()); return ok }) {
		t.Fatal("peers with matching password files did not connect")
	}
}

// An invalid listener option must not leave the socket open.
func TestListenerWithBadOptionsIsClosed(t *testing.T) {
	c := newCoreWith(t)
	probe, err := net.Listen("tcp", "127.0.0.1:0")
	requireNoError(t, err)
	addr := probe.Addr().String()
	_ = probe.Close()
	for _, opt := range []string{"priority=notanumber", "password=a&password_file=/nonexistent"} {
		if _, err := c.Listen(mustParseURL(t, "tcp://"+addr+"?"+opt), ""); err == nil {
			t.Fatalf("%s: listener with invalid options was accepted", opt)
		}
		again, err := net.Listen("tcp", addr)
		if err != nil {
			t.Fatalf("%s: port still held after a rejected listener: %v", opt, err)
		}
		_ = again.Close()
	}
}
