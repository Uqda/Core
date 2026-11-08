// Package autopeer keeps a node connected to healthy public peers.
//
// It watches the node's peer links and, when fewer than the wanted number are
// up, adds the fastest reachable public TLS peers from the public peer
// catalog. It replaces a peer it added when that peer stays down, and swaps
// one of its own peers for a much faster candidate on each periodic check.
// Peers the operator configured are never removed; they count towards the
// target while they are up.
package autopeer

import (
	"context"
	"net"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Uqda/Core/internal/publicpeers"
	"github.com/Uqda/Core/src/core"
)

// Defaults and bounds for the configuration.
const (
	DefaultCount    = 2
	MaxCount        = 5
	DefaultInterval = 30 * time.Minute
	MinInterval     = 10 * time.Minute
	MaxInterval     = 24 * time.Hour
	// DownGrace is how long one of our peers may stay down before it is replaced.
	DownGrace = 5 * time.Minute
	// healthEvery is how often link state is checked locally (no network use).
	healthEvery = time.Minute
	// A candidate must beat a current peer by both margins before a swap,
	// so small timing noise never causes churn.
	swapRatio  = 1.5
	swapMargin = 30 * time.Millisecond
)

// Node is the subset of the Core the manager needs.
type Node interface {
	AddPeer(u *url.URL, sintf string) error
	RemovePeer(u *url.URL, sintf string) error
	GetPeers() []core.PeerInfo
}

// Source returns ranked public peer candidates for a country.
type Source func(ctx context.Context, country string, limit int) ([]publicpeers.Suggestion, error)

// Logger is the subset of the daemon logger the manager uses.
type Logger interface {
	Infof(string, ...interface{})
	Warnf(string, ...interface{})
	Debugf(string, ...interface{})
}

// Config controls the manager.
type Config struct {
	Country  string
	Count    int
	Interval time.Duration
}

type managed struct {
	uri       string
	host      string
	added     time.Time
	downSince time.Time
	connectMS float64
}

// Manager maintains automatic peers. Create it with New and start Run.
type Manager struct {
	cfg    Config
	node   Node
	source Source
	log    Logger
	now    func() time.Time

	mu        sync.Mutex
	peers     map[string]*managed // by host:port
	lastCheck time.Time
	lastErr   string
	failing   bool
}

// New creates a manager. It does nothing until Run is called.
func New(cfg Config, node Node, source Source, log Logger) *Manager {
	if cfg.Count <= 0 {
		cfg.Count = DefaultCount
	}
	if cfg.Interval <= 0 {
		cfg.Interval = DefaultInterval
	}
	return &Manager{cfg: cfg, node: node, source: source, log: log, now: time.Now, peers: map[string]*managed{}}
}

// CatalogSource fetches the public catalog and ranks candidates in it.
func CatalogSource(ctx context.Context, country string, limit int) ([]publicpeers.Suggestion, error) {
	catalog, err := publicpeers.Fetch(ctx)
	if err != nil {
		return nil, err
	}
	report, err := publicpeers.Suggest(ctx, catalog, country, limit)
	if err != nil {
		return nil, err
	}
	return report.Suggestions, nil
}

// Run checks link health every minute and refreshes candidates every
// Interval, until ctx is cancelled.
func (m *Manager) Run(ctx context.Context) {
	m.cycle(ctx, true)
	health := time.NewTicker(healthEvery)
	defer health.Stop()
	refresh := time.NewTicker(m.cfg.Interval)
	defer refresh.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-refresh.C:
			m.cycle(ctx, true)
		case <-health.C:
			m.cycle(ctx, false)
		}
	}
}

// upTransport returns the hosts of established outbound links that reach
// beyond the local network. Inbound links and link-local (multicast) peers
// are not counted: they are not a path to the rest of the network.
func upTransport(links []core.PeerInfo) map[string]bool {
	up := map[string]bool{}
	for _, p := range links {
		if !p.Up || p.Key == nil || p.Inbound {
			continue
		}
		h := hostOf(p.URI)
		if h == "" {
			continue
		}
		if host, _, err := net.SplitHostPort(h); err == nil {
			// Drop any IPv6 zone such as %eth0, which net.ParseIP rejects.
			if i := strings.IndexByte(host, '%'); i >= 0 {
				host = host[:i]
			}
			if ip := net.ParseIP(strings.Trim(host, "[]")); ip != nil && (ip.IsLinkLocalUnicast() || ip.IsLoopback()) {
				continue
			}
		}
		up[h] = true
	}
	return up
}

func hostOf(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	return strings.ToLower(u.Host)
}

// cycle updates health and adds, replaces or swaps peers. refresh allows a
// network fetch for a speed comparison even when nothing is missing.
func (m *Manager) cycle(ctx context.Context, refresh bool) {
	now := m.now()
	links := m.node.GetPeers()
	up := upTransport(links)

	m.mu.Lock()
	// Health of our own peers; drop the ones that stayed down too long.
	var stale []*managed
	for host, p := range m.peers {
		switch {
		case up[host]:
			p.downSince = time.Time{}
		case p.downSince.IsZero():
			p.downSince = now
		case now.Sub(p.downSince) >= DownGrace:
			stale = append(stale, p)
			delete(m.peers, host)
		}
	}
	missing := m.cfg.Count - len(up)
	m.mu.Unlock()

	for _, p := range stale {
		if u, err := url.Parse(p.uri); err == nil {
			_ = m.node.RemovePeer(u, "")
		}
		m.log.Infof("Auto peers: removed %s after %s down", p.uri, DownGrace)
	}
	if missing <= 0 && !refresh {
		return
	}

	candidates, err := m.source(ctx, m.cfg.Country, MaxCount)
	m.mu.Lock()
	m.lastCheck = now
	if err != nil {
		m.lastErr = err.Error()
		first := !m.failing
		m.failing = true
		m.mu.Unlock()
		if first {
			m.log.Warnf("Auto peers: could not refresh public peers for %s: %s (keeping current peers; repeats are logged at debug level)", m.cfg.Country, err)
		} else {
			m.log.Debugf("Auto peers: refresh failed: %s", err)
		}
		return
	}
	m.lastErr, m.failing = "", false

	inUse := map[string]bool{}
	for _, p := range links {
		if h := hostOf(p.URI); h != "" {
			inUse[h] = true
		}
	}
	for h := range m.peers {
		inUse[h] = true
	}
	var free []publicpeers.Suggestion
	byHost := map[string]float64{}
	for _, c := range candidates {
		h := hostOf(c.URI)
		if h == "" {
			continue
		}
		byHost[h] = c.ConnectMS
		if !inUse[h] {
			free = append(free, c)
		}
	}
	sort.SliceStable(free, func(i, j int) bool { return free[i].ConnectMS < free[j].ConnectMS })
	for h, p := range m.peers {
		if ms, ok := byHost[h]; ok {
			p.connectMS = ms
		}
	}

	var toAdd []publicpeers.Suggestion
	var toRemove []*managed
	for missing > 0 && len(free) > 0 {
		toAdd, free = append(toAdd, free[0]), free[1:]
		missing--
	}
	// One swap per refresh at most: our slowest measured peer for a clearly
	// faster free candidate.
	if refresh && len(free) > 0 {
		var slowest *managed
		for _, p := range m.peers {
			if p.connectMS > 0 && (slowest == nil || p.connectMS > slowest.connectMS) {
				slowest = p
			}
		}
		best := free[0]
		if slowest != nil && time.Duration(slowest.connectMS*float64(time.Millisecond)) >
			time.Duration(best.ConnectMS*swapRatio*float64(time.Millisecond))+swapMargin {
			toRemove = append(toRemove, slowest)
			delete(m.peers, slowest.host)
			toAdd = append(toAdd, best)
		}
	}
	for _, c := range toAdd {
		h := hostOf(c.URI)
		m.peers[h] = &managed{uri: c.URI, host: h, added: now, connectMS: c.ConnectMS}
	}
	m.mu.Unlock()

	for _, p := range toRemove {
		if u, err := url.Parse(p.uri); err == nil {
			_ = m.node.RemovePeer(u, "")
		}
		m.log.Infof("Auto peers: replaced %s (%.0f ms) with a faster peer", p.uri, p.connectMS)
	}
	for _, c := range toAdd {
		u, err := url.Parse(c.URI)
		if err != nil {
			continue
		}
		if err := m.node.AddPeer(u, ""); err != nil {
			m.log.Warnf("Auto peers: could not add %s: %s", c.URI, err)
			m.mu.Lock()
			delete(m.peers, hostOf(c.URI))
			m.mu.Unlock()
			continue
		}
		m.log.Infof("Auto peers: added %s (%s, %.0f ms TCP connect)", c.URI, c.Country, c.ConnectMS)
	}
}

// PeerStatus describes one automatic peer.
type PeerStatus struct {
	URI       string  `json:"uri"`
	Up        bool    `json:"up"`
	Added     string  `json:"added"`
	ConnectMS float64 `json:"tcp_connect_ms,omitempty"`
}

// Status is reported by getAutoPeers.
type Status struct {
	Enabled   bool         `json:"enabled"`
	Country   string       `json:"country,omitempty"`
	Count     int          `json:"count,omitempty"`
	Healthy   int          `json:"healthy_peers"`
	Peers     []PeerStatus `json:"auto_peers"`
	LastCheck string       `json:"last_check,omitempty"`
	LastError string       `json:"last_error,omitempty"`
}

// Status reports the manager state. A nil manager reports itself disabled.
func (m *Manager) Status() Status {
	if m == nil {
		return Status{Peers: []PeerStatus{}}
	}
	up := upTransport(m.node.GetPeers())
	m.mu.Lock()
	defer m.mu.Unlock()
	s := Status{Enabled: true, Country: m.cfg.Country, Count: m.cfg.Count, Healthy: len(up), Peers: []PeerStatus{}, LastError: m.lastErr}
	if !m.lastCheck.IsZero() {
		s.LastCheck = m.lastCheck.UTC().Format(time.RFC3339)
	}
	for h, p := range m.peers {
		s.Peers = append(s.Peers, PeerStatus{URI: p.uri, Up: up[h], Added: p.added.UTC().Format(time.RFC3339), ConnectMS: p.connectMS})
	}
	sort.Slice(s.Peers, func(i, j int) bool { return s.Peers[i].URI < s.Peers[j].URI })
	return s
}
