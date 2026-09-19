package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestWriteHostVarsVaultFileIsPrivate(t *testing.T) {
	dir := t.TempDir()
	hostDir := filepath.Join(dir, "host_vars", "01")

	key := keySet{priv: []byte{1, 2, 3, 4}, pub: []byte{5, 6, 7, 8}, ip: "200::1"}
	if err := writeHostVars(hostDir, key); err != nil {
		t.Fatalf("writeHostVars failed: %v", err)
	}

	vaultPath := filepath.Join(hostDir, "vault")
	data, err := os.ReadFile(vaultPath)
	if err != nil {
		t.Fatalf("failed to read vault file: %v", err)
	}
	if !strings.Contains(string(data), "01020304") {
		t.Fatalf("vault file does not contain the expected hex-encoded private key: %s", data)
	}

	// Unix file mode bits aren't meaningful on Windows (os.Chmod there only
	// toggles the read-only attribute), so only assert the exact
	// permission bits on platforms where they mean what we think they mean.
	if runtime.GOOS != "windows" {
		info, err := os.Stat(vaultPath)
		if err != nil {
			t.Fatalf("failed to stat vault file: %v", err)
		}
		if perm := info.Mode().Perm(); perm != 0o600 {
			t.Fatalf("expected vault file to be 0600, got %o", perm)
		}
	}
}

func TestWriteHostVarsFixesExistingLoosePermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix permission bits are not meaningful on Windows")
	}

	dir := t.TempDir()
	hostDir := filepath.Join(dir, "host_vars", "01")
	if err := os.MkdirAll(hostDir, 0755); err != nil {
		t.Fatal(err)
	}
	vaultPath := filepath.Join(hostDir, "vault")
	// Simulate a pre-existing vault file from an older, unfixed version of
	// this tool that left it world-readable.
	if err := os.WriteFile(vaultPath, []byte("stale"), 0644); err != nil {
		t.Fatal(err)
	}

	key := keySet{priv: []byte{9, 9}, pub: []byte{8, 8}, ip: "200::2"}
	if err := writeHostVars(hostDir, key); err != nil {
		t.Fatalf("writeHostVars failed: %v", err)
	}

	info, err := os.Stat(vaultPath)
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Fatalf("expected re-generated vault file to be tightened to 0600, got %o", perm)
	}
}
