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

func TestGroupReadIsAllowedButOtherAccessIsNot(t *testing.T) {
	if !permissionCheckMeaningful {
		t.Skip("permission bits are not meaningful on this platform")
	}
	cases := map[os.FileMode]bool{
		0o600: false, // owner only
		0o640: false, // service group may read (Debian package layout)
		0o660: true,  // group may modify
		0o604: true,  // anyone may read
		0o644: true,
		0o650: true, // group may execute
	}
	for mode, want := range cases {
		if got := filePermissionsAreUnsafe(mode); got != want {
			t.Errorf("mode %o: unsafe=%v, want %v", mode, got, want)
		}
	}
}
