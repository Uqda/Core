package admin

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"
	"time"
)

func healthyDoctorSnapshot() doctorSnapshot {
	return doctorSnapshot{
		buildName:      "uqda",
		buildVersion:   "0.1.3",
		address:        "201::1",
		routingEntries: 4,
		adminListen:    "unix:///var/run/uqda.sock",
		adminMode:      os.FileMode(0600),
		tunKnown:       true,
		tunEnabled:     true,
		tunName:        "utun7",
		tunMTU:         65535,
		peers: []PeerEntry{{
			Up:      true,
			Latency: 38 * time.Millisecond,
		}},
		multicastKnown: true,
		multicastCount: 1,
	}
}

func TestEvaluateDoctorHealthy(t *testing.T) {
	res := evaluateDoctor(healthyDoctorSnapshot())
	if res.Status != DoctorPass {
		t.Fatalf("status = %q, want %q: %#v", res.Status, DoctorPass, res.Checks)
	}
	if len(res.Recommendations) != 0 {
		t.Fatalf("healthy response has recommendations: %#v", res.Recommendations)
	}
}

func TestEvaluateDoctorWithoutBootstrap(t *testing.T) {
	snapshot := healthyDoctorSnapshot()
	snapshot.peers = nil
	snapshot.routingEntries = 1
	snapshot.multicastCount = 0
	res := evaluateDoctor(snapshot)
	if res.Status != DoctorWarn {
		t.Fatalf("status = %q, want %q", res.Status, DoctorWarn)
	}
	assertDoctorCheck(t, res, "peers", DoctorWarn)
	assertDoctorCheck(t, res, "routing", DoctorWarn)
	assertDoctorCheck(t, res, "multicast", DoctorWarn)
}

func TestEvaluateDoctorWithDisconnectedPeers(t *testing.T) {
	snapshot := healthyDoctorSnapshot()
	snapshot.peers = []PeerEntry{{Up: false}, {Up: false}}
	res := evaluateDoctor(snapshot)
	if res.Status != DoctorWarn {
		t.Fatalf("status = %q, want %q", res.Status, DoctorWarn)
	}
	assertDoctorCheck(t, res, "peers", DoctorWarn)
}

func TestEvaluateDoctorDoesNotExposePeerSecrets(t *testing.T) {
	snapshot := healthyDoctorSnapshot()
	snapshot.peers[0].URI = "tls://peer.example:9001?password=topsecret"
	snapshot.peers[0].PublicKey = "private-diagnostic-key"
	res := evaluateDoctor(snapshot)
	encoded, err := json.Marshal(res)
	if err != nil {
		t.Fatal(err)
	}
	for _, secret := range []string{"topsecret", "private-diagnostic-key", "peer.example"} {
		if strings.Contains(string(encoded), secret) {
			t.Fatalf("doctor response exposed %q: %s", secret, encoded)
		}
	}
}

func TestEvaluateDoctorRejectsInsecureAdminEndpoint(t *testing.T) {
	snapshot := healthyDoctorSnapshot()
	snapshot.adminListen = "tcp://0.0.0.0:9001"
	res := evaluateDoctor(snapshot)
	if res.Status != DoctorFail {
		t.Fatalf("status = %q, want %q", res.Status, DoctorFail)
	}
	assertDoctorCheck(t, res, "admin", DoctorFail)
}

func TestEvaluateDoctorAcceptsLoopbackTCPAndWarnsForDisabledTUN(t *testing.T) {
	snapshot := healthyDoctorSnapshot()
	snapshot.adminListen = "tcp://localhost:9001"
	snapshot.tunEnabled = false
	res := evaluateDoctor(snapshot)
	if res.Status != DoctorWarn {
		t.Fatalf("status = %q, want %q", res.Status, DoctorWarn)
	}
	assertDoctorCheck(t, res, "admin", DoctorPass)
	assertDoctorCheck(t, res, "tun", DoctorWarn)
}

func TestEvaluateDoctorFailsWhenUnixSocketCannotBeInspected(t *testing.T) {
	snapshot := healthyDoctorSnapshot()
	snapshot.adminStatError = errors.New("permission denied")
	res := evaluateDoctor(snapshot)
	if res.Status != DoctorFail {
		t.Fatalf("status = %q, want %q", res.Status, DoctorFail)
	}
	assertDoctorCheck(t, res, "admin", DoctorFail)
}

func assertDoctorCheck(t *testing.T, res DoctorResponse, name, status string) {
	t.Helper()
	for _, check := range res.Checks {
		if check.Name == name {
			if check.Status != status {
				t.Fatalf("check %q status = %q, want %q", name, check.Status, status)
			}
			return
		}
	}
	t.Fatalf("check %q not found in %#v", name, res.Checks)
}
