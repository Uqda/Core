package main

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/gologme/log"
)

func TestWarnIfConfigFilePermissionsAreUnsafe(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix permission bits are not meaningful on Windows")
	}

	dir := t.TempDir()

	strict := filepath.Join(dir, "strict.conf")
	if err := os.WriteFile(strict, []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}
	loose := filepath.Join(dir, "loose.conf")
	if err := os.WriteFile(loose, []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}

	newLogger := func() (*log.Logger, *bytes.Buffer) {
		var buf bytes.Buffer
		l := log.New(&buf, "", 0)
		l.EnableLevel("warn")
		return l, &buf
	}

	t.Run("strict permissions produce no warning", func(t *testing.T) {
		logger, buf := newLogger()
		warnIfConfigFilePermissionsAreUnsafe(strict, logger)
		if buf.Len() != 0 {
			t.Fatalf("expected no warning for a 0600 file, got: %s", buf.String())
		}
	})

	t.Run("loose permissions produce a warning naming the path but not any key material", func(t *testing.T) {
		logger, buf := newLogger()
		warnIfConfigFilePermissionsAreUnsafe(loose, logger)
		out := buf.String()
		if !strings.Contains(out, loose) {
			t.Fatalf("expected warning to mention the unsafe path, got: %s", out)
		}
		if !strings.Contains(strings.ToLower(out), "warn") {
			t.Fatalf("expected this to be logged at warn level, got: %s", out)
		}
	})

	t.Run("missing file produces no warning and no panic", func(t *testing.T) {
		logger, buf := newLogger()
		warnIfConfigFilePermissionsAreUnsafe(filepath.Join(dir, "does-not-exist"), logger)
		if buf.Len() != 0 {
			t.Fatalf("expected no warning for a missing file, got: %s", buf.String())
		}
	})
}
