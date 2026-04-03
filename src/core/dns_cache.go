package core

import (
	"context"
	"net"
	"sync"
	"time"
)

func init() {
	go dnsCachePeriodicEviction()
}

// dnsCachePeriodicEviction removes expired entries on a ticker so the map cannot
// retain stale hostnames indefinitely when lookup traffic is low.
func dnsCachePeriodicEviction() {
	t := time.NewTicker(2 * time.Minute)
	defer t.Stop()
	for range t.C {
		globalDNSCache.evictExpired()
	}
}

// dnsCacheEntry stores a DNS lookup result with expiration time.
// err is non-nil when the original LookupIP failed; it must be returned on cache hits
// so callers do not treat a cached failure as success with empty IPs.
type dnsCacheEntry struct {
	ips       []net.IP
	expiresAt time.Time
	err       error
}

// dnsCache provides DNS result caching to reduce lookup latency
type dnsCache struct {
	mu      sync.RWMutex
	entries map[string]*dnsCacheEntry
	// TTL for successful lookups
	successTTL time.Duration
	// TTL for failed lookups (shorter to allow retry)
	failureTTL time.Duration
}

// maxDNSCacheEntries caps hostname cache size so RAM cannot grow without bound if many
// unique hosts are resolved over time (eviction alone only removes expired TTLs).
const maxDNSCacheEntries = 512

var globalDNSCache = &dnsCache{
	entries:    make(map[string]*dnsCacheEntry),
	successTTL: 5 * time.Minute,  // Cache successful lookups for 5 minutes
	failureTTL: 30 * time.Second, // Cache failures for 30 seconds
}

// lookupIPWithCache performs DNS lookup with caching. network must be one of
// "ip", "ip4", or "ip6" (see net.Resolver.LookupIP).
func lookupIPWithCache(hostname, network string) ([]net.IP, error) {
	if network == "" {
		network = "ip"
	}
	cacheKey := hostname + "\x00" + network

	// Check cache first
	globalDNSCache.mu.RLock()
	entry, found := globalDNSCache.entries[cacheKey]
	globalDNSCache.mu.RUnlock()

	if found && time.Now().Before(entry.expiresAt) {
		return entry.ips, entry.err
	}

	// Cache miss or expired - perform actual lookup
	ips, err := net.DefaultResolver.LookupIP(context.Background(), network, hostname)

	// Update cache
	globalDNSCache.mu.Lock()
	defer globalDNSCache.mu.Unlock()

	// Check again in case another goroutine updated it
	if entry, found := globalDNSCache.entries[cacheKey]; found && time.Now().Before(entry.expiresAt) {
		return entry.ips, entry.err
	}

	// Determine TTL based on success/failure
	ttl := globalDNSCache.successTTL
	if err != nil {
		ttl = globalDNSCache.failureTTL
	}

	// Store in cache
	globalDNSCache.entries[cacheKey] = &dnsCacheEntry{
		ips:       ips,
		expiresAt: time.Now().Add(ttl),
		err:       err,
	}

	// Opportunistic sweep when the map grows large (ticker also runs evictExpired).
	if len(globalDNSCache.entries) > 100 {
		globalDNSCache.evictExpired()
	}
	globalDNSCache.trimOverCapacity()

	return ips, err
}

func (c *dnsCache) evictExpired() {
	now := time.Now()
	c.mu.Lock()
	defer c.mu.Unlock()
	for key, entry := range c.entries {
		if now.After(entry.expiresAt) {
			delete(c.entries, key)
		}
	}
	c.trimOverCapacityLocked()
}

// trimOverCapacityLocked removes arbitrary entries if still above maxDNSCacheEntries.
// Caller must hold c.mu (write lock).
func (c *dnsCache) trimOverCapacityLocked() {
	for len(c.entries) > maxDNSCacheEntries {
		for k := range c.entries {
			delete(c.entries, k)
			break
		}
	}
}

func (c *dnsCache) trimOverCapacity() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.trimOverCapacityLocked()
}
