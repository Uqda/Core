package admin

import (
	"slices"
	"testing"
	"time"
)

func TestSortByLatency(t *testing.T) {
	peers := []PeerEntry{
		{URI: "unknown", Up: true},
		{URI: "down", Up: false, Latency: time.Millisecond},
		{URI: "slow", Up: true, Latency: 35 * time.Millisecond, Cost: 35},
		{URI: "fast", Up: true, Latency: 8 * time.Millisecond, Cost: 8},
		{URI: "fast-higher-cost", Up: true, Latency: 8 * time.Millisecond, Cost: 12},
	}

	slices.SortStableFunc(peers, sortByLatency)
	want := []string{"fast", "fast-higher-cost", "slow", "unknown", "down"}
	for i, uri := range want {
		if peers[i].URI != uri {
			t.Fatalf("peer %d = %q, want %q", i, peers[i].URI, uri)
		}
	}
}
