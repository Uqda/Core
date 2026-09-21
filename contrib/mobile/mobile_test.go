package mobile

import (
	"os"
	"strings"
	"testing"

	"github.com/gologme/log"
)

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
