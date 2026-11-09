// Package netwatch notices when the host's network changes, for example when
// Wi-Fi reconnects, a cable is plugged in or an address changes, so peers can
// be redialled at once instead of waiting out their reconnect backoff.
package netwatch

import (
	"context"
	"net"
	"sort"
	"strings"
	"time"
)

// Every is how often interfaces are read. Reading them uses no network.
const Every = 2 * time.Second

// Lister returns a description of the host's usable network state. The
// watcher only compares successive results.
type Lister func() (string, error)

// Watcher calls OnChange once the network state changed and then stayed the
// same for one more reading, so a burst of changes yields one call.
type Watcher struct {
	List     Lister
	Every    time.Duration
	OnChange func()
}

// New watches the host's interfaces.
func New(onChange func()) *Watcher {
	return &Watcher{List: Fingerprint, Every: Every, OnChange: onChange}
}

// Run watches until ctx is cancelled.
func (w *Watcher) Run(ctx context.Context) {
	last, err := w.List()
	if err != nil {
		last = ""
	}
	pending := false
	tick := time.NewTicker(w.Every)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
		}
		cur, err := w.List()
		if err != nil {
			continue
		}
		switch {
		case cur != last:
			last, pending = cur, true
		case pending:
			pending = false
			w.OnChange()
		}
	}
}

// Fingerprint describes the interfaces that are up and their routable
// addresses. Loopback, link-local and Uqda's own addresses (200::/7, which
// also covers 300::/8) are left out: they say nothing about reaching peers.
func Fingerprint() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	var parts []string
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, a := range addrs {
			ipn, ok := a.(*net.IPNet)
			if !ok || !Routable(ipn.IP) {
				continue
			}
			parts = append(parts, iface.Name+"="+ipn.String())
		}
	}
	sort.Strings(parts)
	return strings.Join(parts, ","), nil
}

// Routable reports whether an address can carry peer links.
func Routable(ip net.IP) bool {
	switch {
	case ip == nil, ip.IsLoopback(), ip.IsLinkLocalUnicast(), ip.IsMulticast(), ip.IsUnspecified():
		return false
	case ip.To4() == nil && ip[0]&0xfe == 0x02:
		return false // Uqda overlay
	}
	return true
}
