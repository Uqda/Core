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
