package mobile

import (
	"os"
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

// SendBuffer previously panicked when the caller passed a negative
// length (p[:length] out of range) and Send/SendBuffer with empty
// payload reached writePC, which also panicked.
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
	if err := node.SendBuffer([]byte{1, 2, 3, 4}, -1); err != nil {
		t.Fatalf("SendBuffer returned unexpected error: %s", err)
	}
	if err := node.SendBuffer(nil, 0); err != nil {
		t.Fatalf("SendBuffer returned unexpected error: %s", err)
	}
	if err := node.Send(nil); err != nil {
		t.Fatalf("Send returned unexpected error: %s", err)
	}
}
