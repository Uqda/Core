package admin

import (
	"math"
	"testing"
)

func TestPeerComparatorsDoNotOverflow(t *testing.T) {
	low := PeerEntry{PublicKey: "same", Priority: 0, Cost: 0, Uptime: 0}
	high := PeerEntry{PublicKey: "same", Priority: math.MaxUint64, Cost: math.MaxUint64, Uptime: math.MaxFloat64}

	comparators := map[string]func(PeerEntry, PeerEntry) int{
		"default": sortByDefault,
		"cost":    sortByCost,
		"uptime":  sortByUptime,
	}
	for name, compare := range comparators {
		t.Run(name, func(t *testing.T) {
			if got := compare(low, high); got >= 0 {
				t.Fatalf("low vs high returned %d, want a negative result", got)
			}
			if got := compare(high, low); got <= 0 {
				t.Fatalf("high vs low returned %d, want a positive result", got)
			}
		})
	}
}
