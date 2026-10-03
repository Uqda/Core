package mobile

import (
	"bytes"
	"encoding/json"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gologme/log"
)

func TestStartJSONEnforcesPrivateGroupSessions(t *testing.T) {
	for _, test := range []struct {
		name, remoteGroup string
		allowed           bool
	}{
		{"same group", "mobile-private-test-group", true},
		{"different group", "different-mobile-test-group", false},
		{"public node", "", false},
	} {
		t.Run(test.name, func(t *testing.T) {
			start := func(group string) *Uqda {
				t.Helper()
				configuration, err := json.Marshal(map[string]any{
					"GroupPassword": group, "Peers": []string{}, "Listen": []string{},
					"MulticastInterfaces": []any{}, "IfName": "none",
				})
				if err != nil {
					t.Fatal(err)
				}
				node := &Uqda{}
				if err := node.StartJSON(configuration); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { _ = node.Stop() })
				return node
			}
			host, remote := start("mobile-private-test-group"), start(test.remoteGroup)
			listenerURL, _ := url.Parse("tcp://127.0.0.1:0")
			listener, err := host.core.Listen(listenerURL, "")
			if err != nil {
				t.Fatal(err)
			}
			peerURL, _ := url.Parse("tcp://" + listener.Addr().String())
			if err := remote.core.AddPeer(peerURL, ""); err != nil {
				t.Fatal(err)
			}
			deadline := time.Now().Add(5 * time.Second)
			for {
				peers := remote.core.GetPeers()
				if len(peers) == 1 && peers[0].Up && len(host.core.GetTree()) > 1 && len(remote.core.GetTree()) > 1 {
					break
				}
				if time.Now().After(deadline) {
					t.Fatal("test transport did not connect")
				}
				time.Sleep(10 * time.Millisecond)
			}
			// As in Core's session tests, allow tree/session readiness to converge.
			time.Sleep(3 * time.Second)
			// Reading on the sender processes encrypted session acknowledgments.
			// A transport link must still be possible for non-matching groups.
			_ = host.core.SetReadDeadline(time.Now().Add(8 * time.Second))
			done := make(chan struct{})
			go func() {
				defer close(done)
				buffer := make([]byte, 1024)
				for {
					if _, _, err := host.core.ReadFrom(buffer); err != nil {
						return
					}
				}
			}()
			t.Cleanup(func() {
				_ = host.core.SetReadDeadline(time.Now())
				select {
				case <-done:
				case <-time.After(time.Second):
					t.Error("session reader did not stop")
				}
			})
			payload := []byte("private Umbrel mobile session")
			if _, err := host.core.WriteTo(payload, remote.core.LocalAddr()); err != nil {
				t.Fatal(err)
			}
			timeout := 2 * time.Second
			if test.allowed {
				timeout = 6 * time.Second
			}
			_ = remote.core.SetReadDeadline(time.Now().Add(timeout))
			buffer := make([]byte, 1024)
			n, from, err := remote.core.ReadFrom(buffer)
			if test.allowed {
				if err != nil || !bytes.Equal(buffer[:n], payload) || from.String() != host.core.LocalAddr().String() {
					t.Fatalf("same-group session failed: %v", err)
				}
			} else if err == nil {
				t.Fatal("mobile configuration allowed a non-matching private-group session")
			}
		})
	}
}

func TestStartUqda(t *testing.T) {
	logger := log.New(os.Stdout, "", 0)
	logger.EnableLevel("error")
	logger.EnableLevel("warn")
	logger.EnableLevel("info")

	node := &Uqda{
		logger: logger,
	}
	if err := node.StartAutoconfigure(); err != nil {
		t.Fatalf("Failed to start Uqda: %s", err)
	}
	t.Log("Address:", node.GetAddressString())
	t.Log("Subnet:", node.GetSubnetString())
	t.Log("Routing entries:", node.GetRoutingEntries())
	if err := node.Stop(); err != nil {
		t.Fatalf("Failed to stop Uqda: %s", err)
	}
}

func TestStartJSONRejectsInvalidMulticastConfiguration(t *testing.T) {
	tests := []struct {
		name   string
		config string
		want   string
	}{
		{name: "regex", config: `{MulticastInterfaces: [{Regex: "("}]}`, want: "regex"},
		{name: "priority", config: `{MulticastInterfaces: [{Regex: ".*", Priority: 256}]}`, want: "0-255"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &Uqda{logger: log.New(os.Stdout, "", 0)}
			err := node.StartJSON([]byte(tt.config))
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("expected error containing %q, got %v", tt.want, err)
			}
			if node.core != nil {
				t.Fatal("core started before configuration validation completed")
			}
		})
	}
}

func TestPacketMethodsRejectUnstartedNode(t *testing.T) {
	node := &Uqda{}
	if err := node.Send(nil); err == nil {
		t.Fatal("Send accepted a packet before startup")
	}
	if _, err := node.Recv(); err == nil {
		t.Fatal("Recv read a packet before startup")
	}
}

// TestSendBufferRejectsBadLength verifies malformed buffers return errors
// instead of reaching slice operations or the packet parser.
func TestSendBufferRejectsBadLength(t *testing.T) {
	logger := log.New(os.Stdout, "", 0)
	node := &Uqda{logger: logger}
	if err := node.StartAutoconfigure(); err != nil {
		t.Fatalf("Failed to start Uqda: %s", err)
	}
	defer func() { _ = node.Stop() }()
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("SendBuffer must not panic on bad length, got: %v", r)
		}
	}()
	if err := node.SendBuffer([]byte{1, 2, 3, 4}, -1); err == nil {
		t.Fatal("SendBuffer accepted a negative length")
	}
	if err := node.SendBuffer(nil, 0); err == nil {
		t.Fatal("SendBuffer accepted an empty packet")
	}
	if err := node.Send(nil); err == nil {
		t.Fatal("Send accepted an empty packet")
	}
}
