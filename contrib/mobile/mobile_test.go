package mobile

import (
	"os"
	"testing"

	"github.com/gologme/log"
)

func TestStartUQDA(t *testing.T) {
	logger := log.New(os.Stdout, "", 0)
	logger.EnableLevel("error")
	logger.EnableLevel("warn")
	logger.EnableLevel("info")

	uqda := &UQDA{
		logger: logger,
	}
	if err := uqda.StartAutoconfigure(); err != nil {
		t.Fatalf("Failed to start UQDA: %s", err)
	}
	t.Log("Address:", uqda.GetAddressString())
	t.Log("Subnet:", uqda.GetSubnetString())
	t.Log("Routing entries:", uqda.GetRoutingEntries())
	if err := uqda.Stop(); err != nil {
		t.Fatalf("Failed to stop UQDA: %s", err)
	}
}

// SendBuffer previously panicked when the caller passed a negative
// length (p[:length] out of range) and Send/SendBuffer with empty
// payload reached writePC, which also panicked.
func TestSendBufferRejectsBadLength(t *testing.T) {
	logger := log.New(os.Stdout, "", 0)
	uqda := &UQDA{logger: logger}
	if err := uqda.StartAutoconfigure(); err != nil {
		t.Fatalf("Failed to start UQDA: %s", err)
	}
	defer func() { _ = uqda.Stop() }()
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("SendBuffer must not panic on bad length, got: %v", r)
		}
	}()
	if err := uqda.SendBuffer([]byte{1, 2, 3, 4}, -1); err != nil {
		t.Fatalf("SendBuffer returned unexpected error: %s", err)
	}
	if err := uqda.SendBuffer(nil, 0); err != nil {
		t.Fatalf("SendBuffer returned unexpected error: %s", err)
	}
	if err := uqda.Send(nil); err != nil {
		t.Fatalf("Send returned unexpected error: %s", err)
	}
}
