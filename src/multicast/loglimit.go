package multicast

import (
	"strings"
	"sync"
	"time"
)

// repeatWarnInterval is how long an identical warning stays quiet. Interface
// errors (an address that cannot be bound, a beacon that cannot be sent)
// repeat on every one-second scan; logging each one floods system logs.
const repeatWarnInterval = 10 * time.Minute

// warnLimiter lets the first occurrence of a warning through, then demotes
// identical repeats until repeatWarnInterval has passed.
type warnLimiter struct {
	mu    sync.Mutex
	last  map[string]time.Time
	every time.Duration
	now   func() time.Time
}

func newWarnLimiter() *warnLimiter {
	return &warnLimiter{last: make(map[string]time.Time), every: repeatWarnInterval, now: time.Now}
}

// allow reports whether the warning identified by key should be logged at
// warning level now. The key combines the interface and the error text.
func (w *warnLimiter) allow(key string) bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	now := w.now()
	if t, ok := w.last[key]; ok && now.Sub(t) < w.every {
		return false
	}
	if len(w.last) > 1024 {
		w.last = make(map[string]time.Time)
	}
	w.last[key] = now
	return true
}

// reset forgets warnings for an interface, so that after it recovers a new
// failure is reported immediately.
func (w *warnLimiter) reset(iface string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	prefix := iface + "\x00"
	for k := range w.last {
		if strings.HasPrefix(k, prefix) {
			delete(w.last, k)
		}
	}
}

func warnKey(iface, what string, err error) string {
	return iface + "\x00" + what + "\x00" + err.Error()
}
