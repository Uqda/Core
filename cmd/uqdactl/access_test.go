package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestAdminAccessClassification(t *testing.T) {
	denied := &adminAccessError{cause: fmt.Errorf("dial: %w", os.ErrPermission)}
	if !errors.Is(denied, os.ErrPermission) || !isAdminAccessError(fmt.Errorf("request: %w", denied)) {
		t.Fatal("lost wrapped permission error")
	}
	for _, err := range []error{nil, os.ErrPermission, errors.New("permission denied"), errors.New("connection refused")} {
		if isAdminAccessError(err) {
			t.Fatalf("misclassified unrelated error: %v", err)
		}
	}
}

func TestAdminAccessHints(t *testing.T) {
	for _, platform := range []string{"darwin", "linux"} {
		for command, example := range map[string]string{
			"status": "sudo uqda status", "getSelf": "sudo uqda info",
			"getPeers": "sudo uqda peers", "list": "sudo uqda commands",
			"doctor": "sudo uqdactl doctor", "test": "sudo uqda test IP",
			"getPaths": "sudo uqdactl getPaths", "getSessions": "sudo uqdactl getSessions",
		} {
			if hint := accessHint(command, platform, 501); !strings.Contains(hint, example) {
				t.Errorf("%s %s: %s", platform, command, hint)
			}
		}
		if hint := accessHint("info", platform, 0); strings.Contains(hint, "sudo") || !strings.Contains(hint, "Already running as root") {
			t.Fatalf("wrong root guidance: %s", hint)
		}
	}
	if hint := accessHint("info", "windows", -1); strings.Contains(hint, "sudo uqda") {
		t.Fatalf("suggested sudo on Windows: %s", hint)
	}
	if hint := accessHint("tls://peer?password=TOP_SECRET", "darwin", 501); strings.Contains(hint, "TOP_SECRET") {
		t.Fatal("leaked invocation content")
	}
}
