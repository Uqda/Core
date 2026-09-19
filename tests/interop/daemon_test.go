// Package interop tests the compiled Uqda daemon through configuration files,
// admin sockets and peering. Both processes run the same source revision.
// The independent upstream fixture lives in upstream.py; this Go harness
// remains a separate same-source regression test.
package interop

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/Uqda/Core/src/address"
	"github.com/Uqda/Core/src/admin"
	"github.com/Uqda/Core/src/config"
)

// freeTCPPort asks the OS for an unused loopback TCP port. There is an
// inherent (very small) race between closing this listener and the daemon
// binding the same port, which is an accepted tradeoff for test simplicity;
// it has not been observed to flake in practice.
func freeTCPPort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to allocate a free port: %v", err)
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}

// buildUqda compiles the daemon binary once per test run into a temporary
// directory and returns its path.
func buildUqda(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	binName := "uqda-under-test"
	if os.PathSeparator == '\\' {
		binName += ".exe"
	}
	binPath := filepath.Join(dir, binName)

	cmd := exec.Command("go", "build", "-o", binPath, "github.com/Uqda/Core/cmd/uqda")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to build uqda binary: %v\n%s", err, stderr.String())
	}
	return binPath
}

// nodeSpec is what the caller decides about a node before it exists.
type nodeSpec struct {
	label      string
	listenPort int    // 0 means "do not listen"
	peerAddr   string // "" means "no configured outbound peer"
}

// nodeHandle is what exists once the node has actually been started.
type nodeHandle struct {
	adminPort int
	publicKey []byte
}

// startNode writes a config file for a fresh identity and launches the
// daemon against it. IfName is "none" so the test does not require TUN
// privileges (administrator on Windows, root/CAP_NET_ADMIN on Linux), and
// multicast discovery is disabled so the only path by which the two nodes
// can find each other is the explicit Peers/Listen configuration under
// test.
func startNode(t *testing.T, binPath string, spec nodeSpec) *nodeHandle {
	t.Helper()

	cfg := config.GenerateConfig()
	adminPort := freeTCPPort(t)
	cfg.AdminListen = fmt.Sprintf("tcp://127.0.0.1:%d", adminPort)
	cfg.IfName = "none"
	cfg.MulticastInterfaces = nil
	if spec.listenPort != 0 {
		cfg.Listen = []string{fmt.Sprintf("tcp://127.0.0.1:%d", spec.listenPort)}
	}
	if spec.peerAddr != "" {
		cfg.Peers = []string{spec.peerAddr}
	}

	cfgBytes, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("failed to marshal config for %s: %v", spec.label, err)
	}
	workDir := t.TempDir()
	cfgPath := filepath.Join(workDir, spec.label+".conf")
	if err := os.WriteFile(cfgPath, cfgBytes, 0o600); err != nil {
		t.Fatalf("failed to write config for %s: %v", spec.label, err)
	}

	logPath := filepath.Join(workDir, spec.label+".log")
	logFile, err := os.Create(logPath)
	if err != nil {
		t.Fatalf("failed to create log file for %s: %v", spec.label, err)
	}
	t.Cleanup(func() { logFile.Close() })

	cmd := exec.Command(binPath, "-useconffile", cfgPath, "-loglevel", "debug")
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start %s: %v", spec.label, err)
	}
	label := spec.label
	t.Cleanup(func() {
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		_ = cmd.Wait()
		if t.Failed() {
			if data, err := os.ReadFile(logPath); err == nil {
				t.Logf("--- %s log ---\n%s", label, data)
			}
		}
	})

	privateKey := []byte(cfg.PrivateKey)
	publicKey := ed25519PublicKey(privateKey)

	return &nodeHandle{
		adminPort: adminPort,
		publicKey: publicKey,
	}
}

// ed25519PublicKey extracts the public half from a raw ed25519 private key
// (the last 32 of its 64 bytes), without importing crypto/ed25519 just for
// this one accessor.
func ed25519PublicKey(priv []byte) []byte {
	if len(priv) < 32 {
		return nil
	}
	return priv[len(priv)-32:]
}

// adminCall issues one request against a running node's admin socket and
// returns the raw JSON response payload.
func adminCall(port int, name string) (json.RawMessage, error) {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), 2*time.Second)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	if err := conn.SetDeadline(time.Now().Add(2 * time.Second)); err != nil {
		return nil, err
	}

	if err := json.NewEncoder(conn).Encode(admin.AdminSocketRequest{Name: name}); err != nil {
		return nil, err
	}
	var resp admin.AdminSocketResponse
	if err := json.NewDecoder(conn).Decode(&resp); err != nil {
		return nil, err
	}
	if resp.Status != "success" {
		return nil, fmt.Errorf("admin request %q failed: %s", name, resp.Error)
	}
	return resp.Response, nil
}

func getPeers(port int) (admin.GetPeersResponse, error) {
	var res admin.GetPeersResponse
	raw, err := adminCall(port, "getPeers")
	if err != nil {
		return res, err
	}
	err = json.Unmarshal(raw, &res)
	return res, err
}

func getSelf(port int) (admin.GetSelfResponse, error) {
	var res admin.GetSelfResponse
	raw, err := adminCall(port, "getSelf")
	if err != nil {
		return res, err
	}
	err = json.Unmarshal(raw, &res)
	return res, err
}

// waitForCondition polls fn until it returns true or the timeout elapses,
// then evaluates it one final time so the caller's own failure message
// reflects the true final state rather than a stale poll.
func waitForCondition(timeout time.Duration, fn func() bool) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if fn() {
			return true
		}
		time.Sleep(200 * time.Millisecond)
	}
	return fn()
}

// TestDaemonPeeringAndAddressDerivation runs two independently-configured
// uqda daemon processes, peers them over TCP exactly as a real
// deployment would (via config file + CLI flags, not the internal Go API),
// and checks:
//
//  1. Each daemon's admin socket reports the other as an "up" peer — this
//     validates the whole config-file -> CLI -> core -> transport -> admin
//     path, not just the internal Core API already covered by
//     src/core/core_test.go.
//  2. Each daemon's self-reported IPv6 address matches what an independent
//     computation via src/address.AddrForKey produces from the same public
//     key — the specific invariant a future Uqda build must never break,
//     since any drift here breaks addressing compatibility with the rest
//     of the Yggdrasil network.
func TestDaemonPeeringAndAddressDerivation(t *testing.T) {
	binPath := buildUqda(t)

	listenPort := freeTCPPort(t)
	nodeA := startNode(t, binPath, nodeSpec{label: "nodeA", listenPort: listenPort})
	nodeB := startNode(t, binPath, nodeSpec{
		label:    "nodeB",
		peerAddr: fmt.Sprintf("tcp://127.0.0.1:%d", listenPort),
	})

	connected := waitForCondition(20*time.Second, func() bool {
		pa, errA := getPeers(nodeA.adminPort)
		pb, errB := getPeers(nodeB.adminPort)
		if errA != nil || errB != nil {
			return false
		}
		return len(pa.Peers) == 1 && pa.Peers[0].Up &&
			len(pb.Peers) == 1 && pb.Peers[0].Up
	})
	if !connected {
		pa, errA := getPeers(nodeA.adminPort)
		pb, errB := getPeers(nodeB.adminPort)
		t.Fatalf("nodes did not reach a connected peer state in time: nodeA peers=%+v err=%v; nodeB peers=%+v err=%v",
			pa, errA, pb, errB)
	}

	t.Run("AddressDerivationMatchesNodeA", func(t *testing.T) {
		self, err := getSelf(nodeA.adminPort)
		if err != nil {
			t.Fatalf("getSelf on nodeA failed: %v", err)
		}
		want := net.IP(address.AddrForKey(nodeA.publicKey)[:]).String()
		if self.IPAddress != want {
			t.Fatalf("nodeA address mismatch: daemon reported %q, address.AddrForKey computed %q", self.IPAddress, want)
		}
	})

	t.Run("AddressDerivationMatchesNodeB", func(t *testing.T) {
		self, err := getSelf(nodeB.adminPort)
		if err != nil {
			t.Fatalf("getSelf on nodeB failed: %v", err)
		}
		want := net.IP(address.AddrForKey(nodeB.publicKey)[:]).String()
		if self.IPAddress != want {
			t.Fatalf("nodeB address mismatch: daemon reported %q, address.AddrForKey computed %q", self.IPAddress, want)
		}
	})
}
