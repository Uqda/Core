package admin

import (
	"encoding/binary"
	"testing"
	"time"
)

func TestRecordLookupBoundsAndEvictsHistory(t *testing.T) {
	infos := make(map[lookupKey]lookupInfo)
	now := time.Unix(1_700_000_000, 0)
	var first lookupKey
	for i := 0; i < maxLookupRecords; i++ {
		var key lookupKey
		binary.BigEndian.PutUint64(key[:], uint64(i+1))
		if i == 0 {
			first = key
		}
		recordLookup(infos, key, []uint64{uint64(i)}, now.Add(time.Duration(i)*time.Nanosecond))
	}

	var newest lookupKey
	binary.BigEndian.PutUint64(newest[:], maxLookupRecords+1)
	recordLookup(infos, newest, []uint64{99}, now.Add(maxLookupRecords*time.Nanosecond))
	if len(infos) != maxLookupRecords {
		t.Fatalf("lookup history contains %d records, want %d", len(infos), maxLookupRecords)
	}
	if _, ok := infos[first]; ok {
		t.Fatal("oldest lookup was not evicted")
	}
	if _, ok := infos[newest]; !ok {
		t.Fatal("newest lookup was not recorded")
	}
}

func TestRecordLookupExpiresAndCopiesPath(t *testing.T) {
	infos := make(map[lookupKey]lookupInfo)
	now := time.Unix(1_700_000_000, 0)
	var stale, current lookupKey
	stale[0] = 1
	current[0] = 2
	infos[stale] = lookupInfo{time: now.Add(-lookupRetention - time.Nanosecond)}
	path := []uint64{1, 2, 3}
	recordLookup(infos, current, path, now)
	path[0] = 99

	if _, ok := infos[stale]; ok {
		t.Fatal("expired lookup was not removed")
	}
	if got := infos[current].path[0]; got != 1 {
		t.Fatalf("stored path changed with caller buffer: got %d, want 1", got)
	}
}

func TestRecordLookupBoundsEqualTimestamps(t *testing.T) {
	infos := make(map[lookupKey]lookupInfo)
	now := time.Unix(1_700_000_000, 0)
	for i := 0; i <= maxLookupRecords; i++ {
		var key lookupKey
		binary.BigEndian.PutUint64(key[:], uint64(i+1))
		recordLookup(infos, key, nil, now)
	}
	if len(infos) != maxLookupRecords {
		t.Fatalf("lookup history contains %d records, want %d", len(infos), maxLookupRecords)
	}
}
