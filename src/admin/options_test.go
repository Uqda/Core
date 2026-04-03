package admin

import (
	"encoding/binary"
	"testing"
	"time"
)

func TestPruneLookupTraceMap_dropsStale(t *testing.T) {
	m := make(map[lookupTraceKey]lookupTraceEntry)
	var k lookupTraceKey
	k[0] = 1
	m[k] = lookupTraceEntry{at: time.Now().Add(-2 * time.Hour)}
	pruneLookupTraceMap(m, time.Now())
	if len(m) != 0 {
		t.Fatalf("expected stale entry removed, got %d entries", len(m))
	}
}

func TestPruneLookupTraceMap_keepsFresh(t *testing.T) {
	m := make(map[lookupTraceKey]lookupTraceEntry)
	var k lookupTraceKey
	k[0] = 2
	m[k] = lookupTraceEntry{at: time.Now().Add(-30 * time.Minute)}
	pruneLookupTraceMap(m, time.Now())
	if len(m) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(m))
	}
}

func TestPruneLookupTraceMap_capsSize(t *testing.T) {
	m := make(map[lookupTraceKey]lookupTraceEntry)
	now := time.Now()
	for i := 0; i < maxLookupTraceEntries+50; i++ {
		var k lookupTraceKey
		binary.LittleEndian.PutUint64(k[:8], uint64(i))
		m[k] = lookupTraceEntry{at: now}
	}
	pruneLookupTraceMap(m, now)
	if len(m) > maxLookupTraceEntries {
		t.Fatalf("expected at most %d entries, got %d", maxLookupTraceEntries, len(m))
	}
}
