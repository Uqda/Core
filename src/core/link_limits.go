package core

import (
	"net"
	"sync"
)

// Bounds on inbound connections that have been accepted but have not yet
// completed the metadata handshake. Established peerings are not counted:
// a slot is released as soon as the handshake succeeds or fails.
//
// Each pending handshake costs a goroutine, a deadline timer and, once the
// metadata arrives, an Ed25519 verification. Without a bound, a single host
// can open connections faster than the six-second handshake deadline expires
// them and exhaust memory or CPU.
const (
	maxPendingHandshakes          = 256
	maxPendingHandshakesPerSource = 8
)

// handshakeLimiter counts pending inbound handshakes, globally and per source.
type handshakeLimiter struct {
	mu        sync.Mutex
	total     int
	perSource map[string]int
	maxTotal  int
	maxSource int
}

func newHandshakeLimiter(maxTotal, maxSource int) *handshakeLimiter {
	return &handshakeLimiter{
		perSource: make(map[string]int),
		maxTotal:  maxTotal,
		maxSource: maxSource,
	}
}

// acquire reserves a pending-handshake slot for source. When it returns ok,
// the caller must call release exactly once; release is idempotent.
func (h *handshakeLimiter) acquire(source string) (release func(), ok bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.total >= h.maxTotal || h.perSource[source] >= h.maxSource {
		return nil, false
	}
	h.total++
	h.perSource[source]++
	var once sync.Once
	return func() {
		once.Do(func() {
			h.mu.Lock()
			defer h.mu.Unlock()
			h.total--
			if h.perSource[source]--; h.perSource[source] <= 0 {
				delete(h.perSource, source)
			}
		})
	}, true
}

// handshakeSourceKey groups a remote address for per-source accounting.
// IPv6 sources are grouped by /64, the smallest prefix normally assigned to a
// single site, so rotating addresses inside one prefix does not evade the
// bound. Addresses without an IP (Unix sockets, some wrappers) share a key.
func handshakeSourceKey(addr net.Addr) string {
	if addr == nil {
		return "unknown"
	}
	var ip net.IP
	switch a := addr.(type) {
	case *net.TCPAddr:
		ip = a.IP
	case *net.UDPAddr:
		ip = a.IP
	default:
		host, _, err := net.SplitHostPort(addr.String())
		if err != nil {
			return addr.Network()
		}
		ip = net.ParseIP(host)
		if ip == nil {
			return addr.Network() + ":" + host
		}
	}
	if v4 := ip.To4(); v4 != nil {
		return v4.String()
	}
	return ip.Mask(net.CIDRMask(64, 128)).String() + "/64"
}
