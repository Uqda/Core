//go:build !windows

package main

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/Uqda/Core/src/config"
)

// The quick installer stores uqda.conf as uqda:uqda 0600 and the Debian
// package as root:uqda 0640. Rewriting it must keep the service able to read
// it, without ever granting access to other users.
func TestWriteConfigKeepsOwnerAndGroupPermissions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "uqda.conf")
	if err := os.WriteFile(path, []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, 0o644); err != nil {
		t.Fatal(err)
	}
	wantUID, wantGID := os.Getuid(), os.Getgid()
	if os.Geteuid() == 0 {
		wantUID, wantGID = 1234, 4321
		if err := os.Chown(path, wantUID, wantGID); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := writeConfig(path, config.GenerateConfig(), []byte("{}"), ".bak"); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0o640 {
		t.Fatalf("mode = %o, want 640 (group kept, world access removed)", got)
	}
	st := info.Sys().(*syscall.Stat_t)
	if int(st.Uid) != wantUID || int(st.Gid) != wantGID {
		t.Fatalf("owner = %d:%d, want %d:%d", st.Uid, st.Gid, wantUID, wantGID)
	}
}
