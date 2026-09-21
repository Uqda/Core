package interop

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Uqda/Core/src/config"
)

func runControl(binary, endpoint string, args ...string) ([]byte, error) {
	commandArgs := append([]string{"-endpoint=" + endpoint}, args...)
	return exec.Command(binary, commandArgs...).CombinedOutput()
}

func TestDaemonRejectsInvalidMulticastConfigurationWithoutPanic(t *testing.T) {
	daemon := buildUqda(t)
	cfg := config.GenerateConfig()
	cfg.IfName = "none"
	cfg.AdminListen = "none"
	cfg.MulticastInterfaces = []config.MulticastInterfaceConfig{{Regex: "("}}
	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "invalid.conf")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	output, err := exec.Command(daemon, "-useconffile", path).CombinedOutput()
	if err == nil {
		t.Fatalf("daemon accepted invalid multicast configuration: %s", output)
	}
	text := string(output)
	if strings.Contains(text, "panic:") || !strings.Contains(text, "Configuration is invalid") {
		t.Fatalf("daemon returned an uncontrolled error: %s", output)
	}
}

func TestControlCLIWithLiveDaemon(t *testing.T) {
	daemon := buildUqda(t)
	control := buildCommand(t, "uqdactl", "github.com/Uqda/Core/cmd/uqdactl")
	node := startNode(t, daemon, nodeSpec{label: "control"})
	endpoint := fmt.Sprintf("tcp://127.0.0.1:%d", node.adminPort)
	if !waitForCondition(10*time.Second, func() bool {
		_, err := runControl(control, endpoint, "getSelf")
		return err == nil
	}) {
		t.Fatal("live daemon did not become ready for control requests")
	}

	for _, command := range []string{
		"list", "getSelf", "getPeers", "getTree", "getPaths", "getSessions",
		"getMulticastInterfaces", "getTun",
	} {
		t.Run(command, func(t *testing.T) {
			output, err := runControl(control, endpoint, command)
			if err != nil {
				t.Fatalf("%s failed: %v\n%s", command, err, output)
			}
		})
	}

	t.Run("JSON", func(t *testing.T) {
		output, err := runControl(control, endpoint, "-json", "getSelf")
		if err != nil {
			t.Fatalf("JSON getSelf failed: %v\n%s", err, output)
		}
		if !bytes.Contains(output, []byte(`"build_name"`)) {
			t.Fatalf("JSON output does not contain getSelf data: %s", output)
		}
	})

	t.Run("AddAndRemovePeer", func(t *testing.T) {
		const peer = "tcp://127.0.0.1:1"
		if output, err := runControl(control, endpoint, "addPeer", "uri="+peer); err != nil {
			t.Fatalf("addPeer failed: %v\n%s", err, output)
		}
		if output, err := runControl(control, endpoint, "removePeer", "uri="+peer); err != nil {
			t.Fatalf("removePeer failed: %v\n%s", err, output)
		}
	})

	for name, args := range map[string][]string{
		"UnknownCommand":    {endpoint, "notACommand"},
		"UnsupportedScheme": {"http://127.0.0.1:9001", "getSelf"},
		"UnavailableDaemon": {"tcp://127.0.0.1:0", "getSelf"},
	} {
		t.Run(name, func(t *testing.T) {
			output, err := runControl(control, args[0], args[1:]...)
			if err == nil {
				t.Fatalf("failure path succeeded: %s", output)
			}
			if strings.Contains(string(output), "panic:") {
				t.Fatalf("failure path emitted a panic: %s", output)
			}
		})
	}
}

func TestControlCLIRejectsMalformedResponse(t *testing.T) {
	control := buildCommand(t, "uqdactl", "github.com/Uqda/Core/cmd/uqdactl")
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	done := make(chan error, 1)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			done <- err
			return
		}
		defer conn.Close()
		var request map[string]interface{}
		if err := json.NewDecoder(conn).Decode(&request); err != nil {
			done <- err
			return
		}
		_, err = conn.Write([]byte("{malformed response\n"))
		done <- err
	}()

	endpoint := "tcp://" + listener.Addr().String()
	output, err := runControl(control, endpoint, "getSelf")
	if err == nil {
		t.Fatalf("malformed response succeeded: %s", output)
	}
	if strings.Contains(string(output), "panic:") || !strings.Contains(string(output), "read admin response") {
		t.Fatalf("unexpected malformed-response error: %s", output)
	}
	if serverErr := <-done; serverErr != nil {
		t.Fatalf("malformed response server failed: %v", serverErr)
	}
}
