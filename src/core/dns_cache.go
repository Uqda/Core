package core

import (
	"net"
	"sync"
	"time"
)

// dnsCacheEntry stores a DNS lookup result with expiration time
type dnsCacheEntry struct {
	ips       []net.IP
	expiresAt time.Time
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

var globalDNSCache = &dnsCache{
	entries:    make(map[string]*dnsCacheEntry),
	successTTL: 5 * time.Minute,  // Cache successful lookups for 5 minutes
	failureTTL: 30 * time.Second, // Cache failures for 30 seconds
}

// lookupIPWithCache performs DNS lookup with caching
func lookupIPWithCache(hostname string) ([]net.IP, error) {
	// Check cache first
	globalDNSCache.mu.RLock()
	entry, found := globalDNSCache.entries[hostname]
	globalDNSCache.mu.RUnlock()

	if found && time.Now().Before(entry.expiresAt) {
		// Cache hit - return cached result
		return entry.ips, nil
	}

	// Cache miss or expired - perform actual lookup
	ips, err := net.LookupIP(hostname)

	// Update cache
	globalDNSCache.mu.Lock()
	defer globalDNSCache.mu.Unlock()

	// Check again in case another goroutine updated it
	if entry, found := globalDNSCache.entries[hostname]; found && time.Now().Before(entry.expiresAt) {
		return entry.ips, nil
	}

	// Determine TTL based on success/failure
	ttl := globalDNSCache.successTTL
	if err != nil {
		ttl = globalDNSCache.failureTTL
	}

	// Store in cache
	globalDNSCache.entries[hostname] = &dnsCacheEntry{
		ips:       ips,
		expiresAt: time.Now().Add(ttl),
	}

	// Cleanup old entries periodically (simple cleanup every 100 lookups)
	if len(globalDNSCache.entries) > 100 {
		now := time.Now()
		for host, entry := range globalDNSCache.entries {
			if now.After(entry.expiresAt) {
				delete(globalDNSCache.entries, host)
			}
		}
	}

	return ips, err
}
