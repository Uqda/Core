package interop

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/Uqda/Core/src/version"
)

func TestReleaseIdentity(t *testing.T) {
	source, err := os.ReadFile("../../src/version/VERSION")
	if err != nil {
		t.Fatal(err)
	}
	machine := strings.TrimSpace(string(source))
	if machine != version.BuildVersion() {
		t.Fatal("embedded machine version differs from VERSION")
	}
	parts := regexp.MustCompile(`^(\d{2})\.(\d+)(?:-beta\.(\d+))?$`).FindStringSubmatch(machine)
	if parts == nil {
		t.Fatalf("invalid annual version %q", machine)
	}
	want := "Uqda Core " + parts[1]
	if parts[2] != "0" {
		want += "." + parts[2]
	}
	if parts[3] != "" {
		want += " Beta " + parts[3]
	}
	if version.DisplayName() != want {
		t.Fatalf("display name: got %q, want %q", version.DisplayName(), want)
	}
	for _, command := range []struct{ name, arg string }{{"uqda", "--version"}, {"uqdactl", "version"}} {
		t.Run(command.name, func(t *testing.T) {
			binary := filepath.Join(t.TempDir(), command.name)
			if os.PathSeparator == '\\' {
				binary += ".exe"
			}
			cmd := exec.Command("go", "build", "-o", binary, "github.com/Uqda/Core/cmd/"+command.name)
			if output, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("build: %v\n%s", err, output)
			}
			output, err := exec.Command(binary, command.arg).CombinedOutput()
			if err != nil {
				t.Fatalf("version: %v: %s", err, output)
			}
			if string(output) != fmt.Sprintln(want) {
				t.Fatalf("got %q, want %q", output, fmt.Sprintln(want))
			}
		})
	}
}
