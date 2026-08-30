//go:build aix || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris

package main

import (
	"os"
	"testing"
)

func TestRequireAdministratorMatchesEffectiveUser(t *testing.T) {
	err := requireAdministrator()
	if os.Geteuid() == 0 && err != nil {
		t.Fatalf("root was rejected: %v", err)
	}
	if os.Geteuid() != 0 && err == nil {
		t.Fatal("non-root user was allowed")
	}
}
