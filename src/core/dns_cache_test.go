package core

import (
	"strconv"
	"testing"
	"time"
)

func TestDNSCache_evictExpired(t *testing.T) {
	c := &dnsCache{
		entries:    make(map[string]*dnsCacheEntry),
		successTTL: time.Minute,
		failureTTL: time.Second,
	}
	c.entries["stale"] = &dnsCacheEntry{expiresAt: time.Now().Add(-time.Hour)}
	c.entries["fresh"] = &dnsCacheEntry{expiresAt: time.Now().Add(time.Hour)}
	c.evictExpired()
	if len(c.entries) != 1 {
		t.Fatalf("expected 1 entry after eviction, got %d", len(c.entries))
	}
	if _, ok := c.entries["fresh"]; !ok {
		t.Fatal("expected fresh entry to remain")
	}
}

func TestDNSCache_trimOverCapacity(t *testing.T) {
	c := &dnsCache{
		entries:    make(map[string]*dnsCacheEntry),
		successTTL: time.Hour,
		failureTTL: time.Minute,
	}
	future := time.Now().Add(time.Hour)
	for i := 0; i < maxDNSCacheEntries+20; i++ {
		key := strconv.Itoa(i) + "\x00ip"
		c.entries[key] = &dnsCacheEntry{expiresAt: future}
	}
	c.trimOverCapacity()
	if len(c.entries) > maxDNSCacheEntries {
		t.Fatalf("expected at most %d entries, got %d", maxDNSCacheEntries, len(c.entries))
	}
}
