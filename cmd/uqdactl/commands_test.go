package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/Uqda/Core/src/admin"
)

func TestCommandLineOptionsAnywhereAndAliases(t *testing.T) {
	for _, args := range [][]string{
		{"--json", "peers", "--endpoint=tcp://127.0.0.1:1234"},
		{"peers", "-endpoint", "tcp://127.0.0.1:1234", "-json"},
	} {
		env := newCmdLineEnv()
		if err := env.parseFlagsAndArgs(args, &bytes.Buffer{}); err != nil {
			t.Fatal(err)
		}
		if !env.injson || env.server != "tcp://127.0.0.1:1234" || env.args[0] != "getpeers" {
			t.Fatalf("options were ignored: %+v", env)
		}
	}
	env := newCmdLineEnv()
	if err := env.parseFlagsAndArgs(nil, &bytes.Buffer{}); err != nil || env.args[0] != "status" {
		t.Fatalf("default must be status: %+v, %v", env, err)
	}
}

func TestCommandLineRejectsIgnoredArguments(t *testing.T) {
	for _, args := range [][]string{
		{"status", "unused"}, {"version", "extra"}, {"peers", "unused"},
		{"peers", "sort=a", "sort=b"}, {"peers", "=value"},
		{"status", "--does-not-exist"}, {"status", "--endpoint"},
		{"--version", "peers"},
	} {
		env := newCmdLineEnv()
		if err := env.parseFlagsAndArgs(args, &bytes.Buffer{}); err == nil {
			t.Fatalf("accepted ignored arguments: %v", args)
		}
	}
}

func TestAdminArgumentsMatchRunningNode(t *testing.T) {
	available := admin.ListResponse{List: []admin.ListEntry{{Command: "getpeers", Fields: []string{"sort"}}}}
	if err := validateAdminArguments([]string{"getPeers", "sort=uptime"}, available); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"getPeers", "unused=value"}, {"nonexistent"}} {
		if err := validateAdminArguments(args, available); err == nil {
			t.Fatalf("accepted unavailable command/argument: %v", args)
		}
	}
}

func TestPeerOutputDoesNotLeakCredentials(t *testing.T) {
	var report admin.GetPeersResponse
	// Populate through JSON to exercise the actual admin field names.
	endpoint, stop := startDoctorServer(t, map[string]interface{}{"getPeers": map[string]interface{}{
		"peers": []interface{}{map[string]interface{}{
			"remote":     "tls://user:private-secret@example.org:1234?password=private-secret",
			"last_error": "password=private-secret", "up": false,
		}},
	}})
	defer stop()
	if err := doctorRequest(endpoint, "getPeers", &report); err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	renderPeers(&output, report)
	if strings.Contains(output.String(), "private-secret") || strings.Contains(output.String(), "user:") {
		t.Fatal("peer output leaked authentication data")
	}
	if !strings.Contains(output.String(), "[DOWN]") || !strings.Contains(output.String(), "example.org:1234") {
		t.Fatalf("missing useful peer status: %s", output.String())
	}
}

func TestNetworkOutputShowsLossWithoutClaimingRawRTT(t *testing.T) {
	var output bytes.Buffer
	renderNetworkTest(&output, networkTestReport{
		Target: "200:1234::1", Status: "fail", Sent: 3, Received: 0,
		LossPercent: 100, MeasurementNote: "Includes ping process startup; not raw network RTT.",
	})
	for _, expected := range []string{"NETWORK TEST", "[FAIL]", "100% loss", "no reply", "not raw network RTT"} {
		if !strings.Contains(output.String(), expected) {
			t.Fatalf("missing %q: %s", expected, output.String())
		}
	}
}
