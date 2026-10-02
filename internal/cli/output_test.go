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
			if strings.Contains(line, "Connected peers") || strings.Contains(line, "Installed release") {
				if len(line) < 23 || line[23] == ' ' {
					t.Fatalf("misaligned command description: %q", line)
				}
			}
		}
	}
}
