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

	uqda := &Uqda{
		logger: logger,
	}
	if err := uqda.StartAutoconfigure(); err != nil {
		t.Fatalf("Failed to start Uqda: %s", err)
	}
	t.Log("Address:", uqda.GetAddressString())
	t.Log("Subnet:", uqda.GetSubnetString())
	t.Log("Routing entries:", uqda.GetRoutingEntries())
	if err := uqda.Stop(); err != nil {
		t.Fatalf("Failed to stop Uqda: %s", err)
	}
}
