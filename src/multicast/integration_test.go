package multicast

import "testing"

// TestMulticastDiscovery_TwoNodes_Lab is a placeholder for a future integration
// test that boots two nodes on the same L2 segment and asserts discovery via
// ff02::114. Default CI environments do not provide reliable multicast between
// test processes; run manually in a lab when changing discovery logic.
func TestMulticastDiscovery_TwoNodes_Lab(t *testing.T) {
	t.Skip("lab-only: two uqda instances on the same LAN with multicast enabled; not run in default CI")
}
