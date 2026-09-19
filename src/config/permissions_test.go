package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFilePermissionsAreUnsafe(t *testing.T) {
	dir := t.TempDir()

	strict := filepath.Join(dir, "strict.conf")
	if err := os.WriteFile(strict, []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}
	loose := filepath.Join(dir, "loose.conf")
	if err := os.WriteFile(loose, []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}

	unsafeStrict, ok, err := FilePermissionsAreUnsafe(strict)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok != permissionCheckMeaningful {
		t.Fatalf("ok=%v, want %v (permissionCheckMeaningful)", ok, permissionCheckMeaningful)
	}
	if ok && unsafeStrict {
		t.Fatalf("0600 file should not be reported unsafe")
	}

	unsafeLoose, ok, err := FilePermissionsAreUnsafe(loose)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok && !unsafeLoose {
		t.Fatalf("0644 file should be reported unsafe on platforms where the check is meaningful")
	}
	if !ok && unsafeLoose {
		t.Fatalf("unsafe should be false when the check is not meaningful (ok=false), to avoid a false-positive warning")
	}
}

func TestFilePermissionsAreUnsafeMissingFile(t *testing.T) {
	_, _, err := FilePermissionsAreUnsafe(filepath.Join(t.TempDir(), "does-not-exist"))
	if err == nil {
		t.Fatal("expected an error for a missing file")
	}
}
