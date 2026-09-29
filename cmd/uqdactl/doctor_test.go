package main

import (
	"encoding/json"
	"net"
	"strings"
	"testing"

	"github.com/Uqda/Core/src/admin"
)

func TestDoctorCommandAliases(t *testing.T) {
	for _, name := range []string{"doctor", "status", "DOCTOR"} {
		if !isDoctorCommand(name) {
			t.Fatalf("%q should be a doctor alias", name)
		}
	}
	if isDoctorCommand("getPeers") {
		t.Fatal("existing admin commands must not be intercepted")
	}
}

func TestDoctorReportsHealthyNodeWithoutSensitivePeerData(t *testing.T) {
	endpoint, stop := startDoctorServer(t, map[string]interface{}{
		"getSelf": map[string]interface{}{
			"key": strings.Repeat("a", 64), "address": "200:1234::1",
		},
		"getPeers": map[string]interface{}{
			"peers": []interface{}{map[string]interface{}{
				"up": true, "remote": "tls://node.example?password=secret", "last_error": "secret",
			}},
		},
		"getTun": map[string]interface{}{"enabled": true, "name": "uqda0"},
	})
	defer stop()
	report := diagnoseNode(endpoint)
	if len(report.Checks) != 4 {
		t.Fatalf("expected four checks, got %+v", report.Checks)
	}
	for _, check := range report.Checks {
		if check.Status != "pass" {
			t.Fatalf("unexpected check: %+v", check)
		}
	}
	encoded, err := json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), "secret") || strings.Contains(string(encoded), "node.example") {
		t.Fatalf("doctor report leaked peer details: %s", encoded)
	}
}

func TestDoctorReportsNoPeersAndRouterOnly(t *testing.T) {
	endpoint, stop := startDoctorServer(t, map[string]interface{}{
		"getSelf":  map[string]interface{}{"key": strings.Repeat("b", 64), "address": "200:1234::1"},
		"getPeers": map[string]interface{}{"peers": []interface{}{}},
		"getTun":   map[string]interface{}{"enabled": false},
	})
	defer stop()
	report := diagnoseNode(endpoint)
	if report.Checks[2].Status != "warn" || report.Checks[3].Status != "warn" {
		t.Fatalf("expected actionable warnings: %+v", report.Checks)
	}
}

func TestDoctorRetriesWhileAdminSocketStarts(t *testing.T) {
	attempts := 0
	endpoint, stop := startDoctorServer(t, map[string]interface{}{
		"getSelf": func() interface{} {
			attempts++
			if attempts < 3 {
				return nil // Simulate a listener that has not initialized its handler.
			}
			return map[string]interface{}{"key": strings.Repeat("c", 64), "address": "200:1234::1"}
		},
		"getPeers": map[string]interface{}{"peers": []interface{}{}},
		"getTun":   map[string]interface{}{"enabled": true},
	})
	defer stop()
	report := diagnoseNode(endpoint)
	if attempts != 3 || report.Checks[0].Status != "pass" {
		t.Fatalf("doctor did not tolerate startup delay: attempts=%d checks=%+v", attempts, report.Checks)
	}
}

func TestDoctorReportsUnavailableDaemon(t *testing.T) {
	report := diagnoseNode("tcp://127.0.0.1:0")
	if len(report.Checks) != 1 || report.Checks[0].Status != "fail" {
		t.Fatalf("expected daemon failure: %+v", report.Checks)
	}
}

func TestDoctorWarnsAboutNonlocalAdminEndpoint(t *testing.T) {
	report := diagnoseNode("tcp://192.0.2.1:0")
	if len(report.Checks) < 1 || report.Checks[0].Name != "Admin" || report.Checks[0].Status != "warn" {
		t.Fatalf("expected nonlocal admin warning: %+v", report.Checks)
	}
}

func startDoctorServer(t *testing.T, replies map[string]interface{}) (string, func()) {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			func() {
				defer conn.Close()
				var request admin.AdminSocketRequest
				if err := json.NewDecoder(conn).Decode(&request); err != nil {
					return
				}
				payload, ok := replies[request.Name]
				if !ok {
					return
				}
				if responder, ok := payload.(func() interface{}); ok {
					payload = responder()
					if payload == nil {
						return
					}
				}
				response, _ := json.Marshal(payload)
				_ = json.NewEncoder(conn).Encode(admin.AdminSocketResponse{Status: "success", Response: response})
			}()
		}
	}()
	return "tcp://" + listener.Addr().String(), func() {
		_ = listener.Close()
		<-done
	}
}
