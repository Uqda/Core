//go:build aix || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris

package admin

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSecureUnixSocketIsOwnerOnly(t *testing.T) {
	path := filepath.Join(t.TempDir(), "uqda.sock")
	if err := os.WriteFile(path, nil, 0666); err != nil {
		t.Fatal(err)
	}
	if err := secureUnixSocket(path); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0600 {
		t.Fatalf("permissions = %#o, want 0600", got)
	}
}
