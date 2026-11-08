package interop

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/Uqda/Core/src/config"
)

// TestInviteAndJoinWithLiveDaemon issues an invite from a running daemon and
// redeems it with the real uqdactl binary into a brand-new configuration.
func TestInviteAndJoinWithLiveDaemon(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("redemption uses the host's LAN address; exercised on Linux runners")
	}
	daemon := buildUqda(t)
	control := buildCommand(t, "uqdactl", "github.com/Uqda/Core/cmd/uqdactl")
	const group = "interop-invite-group-secret-value"
	node := startNode(t, daemon, nodeSpec{label: "inviter", group: group, listenPort: freeTCPPort(t)})
	endpoint := fmt.Sprintf("tcp://127.0.0.1:%d", node.adminPort)
	if !waitForCondition(10*time.Second, func() bool {
		_, err := runControl(control, endpoint, "getSelf")
		return err == nil
	}) {
		t.Fatal("daemon did not become ready")
	}

	output, err := runControl(control, endpoint, "-json", "createInvite", "expires=2m", "name=interop")
	if err != nil {
		if bytes.Contains(output, []byte("no usable network address")) {
			t.Skip("runner has no non-loopback address")
		}
		t.Fatalf("createInvite failed: %v\n%s", err, output)
	}
	var created struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(output, &created); err != nil || !strings.HasPrefix(created.Token, "uqda-invite-1:") {
		t.Fatalf("unexpected createInvite output: %s", output)
	}
	if bytes.Contains(output, []byte(group)) {
		t.Fatal("createInvite output leaks the group password")
	}

	path := filepath.Join(t.TempDir(), "joined.conf")
	joinOut, err := exec.Command(control, "-json", "join", created.Token, "config="+path).CombinedOutput()
	if err != nil {
		t.Fatalf("join failed: %v\n%s", err, joinOut)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	joined := config.GenerateConfig()
	if _, err := joined.ReadFrom(bytes.NewReader(data)); err != nil {
		t.Fatalf("joined configuration does not parse: %v", err)
	}
	if joined.GroupPassword != group {
		t.Fatal("joined configuration did not receive the group password")
	}
	if !joined.Firewall.Enabled {
		t.Fatal("a new configuration from join should have the firewall on")
	}
	if info, err := os.Stat(path); err == nil && info.Mode().Perm()&0o077 != 0 {
		t.Fatalf("joined configuration is readable by others: %v", info.Mode())
	}

	// The invite was single-use.
	again, err := exec.Command(control, "join", created.Token, "config="+filepath.Join(t.TempDir(), "x.conf")).CombinedOutput()
	if err == nil {
		t.Fatalf("a used invite worked a second time: %s", again)
	}
}
