//go:build linux || darwin

package main

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestDeniedUnixSocketCommands(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("requires an unprivileged user: root bypasses Unix socket mode bits")
	}
	dir, err := os.MkdirTemp("", "uqda-access-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	path := filepath.Join(dir, "s")
	listener, err := net.Listen("unix", path)
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	if err := os.Chmod(path, 0); err != nil {
		t.Fatal(err)
	}
	defer os.Chmod(path, 0600)
	endpoint := "unix://" + path
	_, err = dialAdminEndpoint(endpoint, log.New(io.Discard, "", 0))
	if !isAdminAccessError(err) || !errors.Is(err, os.ErrPermission) {
		t.Fatalf("expected typed permission denial, got %v", err)
	}
	for _, args := range [][]string{
		{"status"}, {"status", "--json"}, {"doctor"}, {"info"}, {"peers"}, {"commands"},
		{"getPaths"}, {"getSessions"}, {"test", "200:1::1"},
		{"addPeer", "uri=tls://peer.example:1234?password=TOP_SECRET"},
	} {
		t.Run(strings.Join(args[:1], ""), func(t *testing.T) {
			originalArgs, originalOut, originalErr := os.Args, os.Stdout, os.Stderr
			read, write, err := os.Pipe()
			if err != nil {
				t.Fatal(err)
			}
			defer read.Close()
			defer func() { os.Args, os.Stdout, os.Stderr = originalArgs, originalOut, originalErr }()
			os.Args = append([]string{"uqdactl", "--endpoint", endpoint}, args...)
			os.Stdout, os.Stderr = write, write
			start := time.Now()
			code := run()
			write.Close()
			output, err := io.ReadAll(read)
			if err != nil {
				t.Fatal(err)
			}
			if code != 1 || !strings.Contains(string(output), "sudo") || strings.Contains(string(output), "TOP_SECRET") {
				t.Fatalf("exit %d, output %s", code, output)
			}
			if time.Since(start) > 2*time.Second {
				t.Fatal("retried a permanent permission denial")
			}
			if len(args) == 2 && args[1] == "--json" {
				var report doctorReport
				if err := json.Unmarshal(output, &report); err != nil || len(report.Checks) != 1 || report.Checks[0].Name != "Access" {
					t.Fatalf("invalid permission report: %s (%v)", output, err)
				}
			}
		})
	}
}
