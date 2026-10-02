package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestHelpIsAlignedAndPortable(t *testing.T) {
	for _, program := range []string{"uqda", "uqdactl"} {
		var output bytes.Buffer
		Help(&output, program, "26.0.1")
		text := output.String()
		for _, char := range text {
			if char > 127 || char == '\x1b' {
				t.Fatal("output requires Unicode or terminal escape support")
			}
		}
		for _, line := range strings.Split(text, "\n") {
			if len(line) > 80 {
				t.Fatalf("help exceeds terminal width: %q", line)
			}
			if strings.Contains(line, "Connected peers") || strings.Contains(line, "Installed release") {
				if len(line) < 23 || line[23] == ' ' {
					t.Fatalf("misaligned command description: %q", line)
				}
			}
		}
	}
}

func TestCompleteHelpAndTopics(t *testing.T) {
	var output bytes.Buffer
	Help(&output, "uqda", "test")
	for _, entry := range append(append([]guideEntry{}, commandGuide...), lifecycleGuide...) {
		if !strings.Contains(output.String(), entry.syntax) {
			t.Fatalf("missing guide entry %s", entry.name)
		}
		var topic bytes.Buffer
		if !HelpTopic(&topic, "uqda", "test", entry.name) || !strings.Contains(topic.String(), entry.syntax) {
			t.Fatalf("missing command help %s", entry.name)
		}
	}
	for _, alias := range []string{"doctor", "getself", "GETPEERS", "list"} {
		if !HelpTopic(&output, "uqdactl", "test", alias) {
			t.Fatalf("missing alias help %s", alias)
		}
	}
	if HelpTopic(&output, "uqda", "test", "not-a-command") {
		t.Fatal("unknown help topic accepted")
	}
	for _, warning := range []string{"runtime state", "private key", "GroupPassword", "Exit codes", "--count=10"} {
		if !strings.Contains(output.String(), warning) {
			t.Fatalf("missing usage or safety note %s", warning)
		}
	}
}
