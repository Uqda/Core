// Package autopeer keeps a node connected to healthy public peers.
//
// It watches the node's peer links every few seconds and, when fewer than the
// wanted number are up, adds the next peer from a standby list that was
// measured in advance, so a lost peer is replaced without waiting for a
// network fetch. A peer that does not come up in time is dropped and the next
// candidate is tried. Peers it added are removed when they stay down, when
// they are the slowest of more peers than wanted, or when a much faster
// candidate appears. Peers the operator configured are never removed; they
// count towards the target while they are up.
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
	// DownGrace is how long a peer we added may stay down before it is
	// removed. A replacement is added at once; this only decides when the
	// old link stops being retried.
	DownGrace = 5 * time.Minute
)

// Timing of the local checks. Checks read link state only and use no network.
const (
	// CheckEvery is how often link state is checked.
	CheckEvery = 2 * time.Second
	// UpWait is how long a newly added peer has to come up. After that it
	// is removed and the next candidate is tried.
	UpWait = 20 * time.Second
	// FailedFor is how long a peer that never came up is skipped.
	FailedFor = 10 * time.Minute
	// StableFor is how long peers must be up before an extra one is trimmed,
	// so a flapping link is not traded for another.
	StableFor = 30 * time.Second
	// standbySize is how many measured spare candidates are kept ready.
	standbySize = 3
	// minFetchGap limits catalog fetches triggered by an empty standby list.
	minFetchGap = time.Minute
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
	everUp    bool
	upSince   time.Time
	downSince time.Time
	connectMS float64       // TCP connect time measured from the catalog
	latency   time.Duration // round trip measured on the live link
}

// speed is the best known round-trip estimate for a peer.
func (p *managed) speed() time.Duration {
	if p.latency > 0 {
		return p.latency
	}
	return time.Duration(p.connectMS * float64(time.Millisecond))
}

// Manager maintains automatic peers. Create it with New and start Run.
type Manager struct {
	cfg    Config
	node   Node
	source Source
	log    Logger
	now    func() time.Time
	kick   chan struct{}

	mu        sync.Mutex
	peers     map[string]*managed // by host:port
	standby   []publicpeers.Suggestion
	failed    map[string]time.Time // host -> skip until
	fetching  bool
	lastFetch time.Time
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
	return &Manager{
		cfg: cfg, node: node, source: source, log: log, now: time.Now,
		kick:  make(chan struct{}, 1),
		peers: map[string]*managed{}, failed: map[string]time.Time{},
	}
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

// Kick asks for an immediate check and a fresh measurement of candidates,
// for example after the host's network changed. It never blocks.
func (m *Manager) Kick() {
	if m == nil {
		return
	}
	select {
	case m.kick <- struct{}{}:
	default:
	}
}

// Run checks link state every CheckEvery, measures candidates every
// Interval, and reacts at once to Kick, until ctx is cancelled.
func (m *Manager) Run(ctx context.Context) {
	m.refresh(ctx)
	m.check()
	tick := time.NewTicker(CheckEvery)
	defer tick.Stop()
	refresh := time.NewTicker(m.cfg.Interval)
	defer refresh.Stop()
	done := make(chan struct{}, 1)
	fetch := func(force bool) {
		m.mu.Lock()
		if m.fetching || (!force && m.now().Sub(m.lastFetch) < minFetchGap) {
			m.mu.Unlock()
			return
		}
		m.fetching = true
		m.mu.Unlock()
		go func() {
			m.refresh(ctx)
			m.mu.Lock()
			m.fetching = false
			m.mu.Unlock()
			select {
			case done <- struct{}{}:
			default:
			}
		}()
	}
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			if m.check() {
				fetch(false)
			}
		case <-done:
			m.check()
		case <-refresh.C:
			fetch(true)
		case <-m.kick:
			m.check()
			fetch(true)
		}
	}
}

// upTransport returns the established outbound links that reach beyond the
// local network, by host. Inbound links and link-local (multicast) peers are
// not counted: they are not a path to the rest of the network.
func upTransport(links []core.PeerInfo) map[string]core.PeerInfo {
	up := map[string]core.PeerInfo{}
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
		up[h] = p
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

// check reads link state, replaces lost peers from the standby list, drops
// peers that never came up or stayed down, and trims the slowest extra peer.
// It uses no network. It reports whether the standby list ran out while a
// peer was still missing, so a fetch is needed.
func (m *Manager) check() (needFetch bool) {
	now := m.now()
	links := m.node.GetPeers()
	up := upTransport(links)
	inUse := map[string]bool{}
	for _, p := range links {
		if h := hostOf(p.URI); h != "" {
			inUse[h] = true
		}
	}

	m.mu.Lock()
	m.lastCheck = now
	var drop []*managed
	var reasons []string
	pending := 0
	for host, p := range m.peers {
		info, isUp := up[host]
		switch {
		case isUp:
			if !p.everUp || !p.downSince.IsZero() || p.upSince.IsZero() {
				p.upSince = now
			}
			p.everUp, p.downSince = true, time.Time{}
			if info.Latency > 0 {
				p.latency = info.Latency
			}
		case !p.everUp && now.Sub(p.added) >= UpWait:
			drop, reasons = append(drop, p), append(reasons, "did not come up within "+UpWait.String())
			m.failed[host] = now.Add(FailedFor)
			delete(m.peers, host)
		case !p.everUp:
			pending++
		case p.downSince.IsZero():
			p.downSince, p.upSince = now, time.Time{}
		case now.Sub(p.downSince) >= DownGrace:
			drop, reasons = append(drop, p), append(reasons, "down for "+DownGrace.String())
			delete(m.peers, host)
		}
	}
	for host, until := range m.failed {
		if !now.Before(until) {
			delete(m.failed, host)
		}
	}

	// Replace what is missing from the measured standby list.
	missing := m.cfg.Count - len(up) - pending
	var add []publicpeers.Suggestion
	for missing > 0 && len(m.standby) > 0 {
		c := m.standby[0]
		m.standby = m.standby[1:]
		h := hostOf(c.URI)
		if h == "" || inUse[h] || m.peers[h] != nil || !m.failed[h].IsZero() {
			continue
		}
		add = append(add, c)
		inUse[h] = true
		m.peers[h] = &managed{uri: c.URI, host: h, added: now, connectMS: c.ConnectMS}
		missing--
	}
	// Fetch when a peer is still missing, or to refill the standby list right
	// after it was used, so the next failure is just as quick. A healthy node
	// never fetches here; the periodic refresh keeps the list current.
	needFetch = missing > 0 || (len(add) > 0 && len(m.standby) < 2)

	// More peers up than wanted: drop our slowest one, once all are stable.
	if missing <= 0 && len(add) == 0 && len(up) > m.cfg.Count {
		var slowest *managed
		stable := true
		for host, p := range m.peers {
			if _, isUp := up[host]; !isUp {
				continue
			}
			if p.upSince.IsZero() || now.Sub(p.upSince) < StableFor {
				stable = false
				break
			}
			if slowest == nil || p.speed() > slowest.speed() {
				slowest = p
			}
		}
		if stable && slowest != nil {
			drop, reasons = append(drop, slowest), append(reasons, "slowest of more peers than wanted")
			delete(m.peers, slowest.host)
		}
	}
	m.mu.Unlock()

	for i, p := range drop {
		if u, err := url.Parse(p.uri); err == nil {
			_ = m.node.RemovePeer(u, "")
		}
		m.log.Infof("Auto peers: removed %s (%s)", p.uri, reasons[i])
	}
	m.addPeers(add, "replacing a lost peer")
	return needFetch
}

// addPeers connects to new peers; failures are forgotten so the next
// candidate is tried on the next check.
func (m *Manager) addPeers(add []publicpeers.Suggestion, why string) {
	for _, c := range add {
		u, err := url.Parse(c.URI)
		if err == nil {
			err = m.node.AddPeer(u, "")
		}
		if err != nil {
			m.log.Warnf("Auto peers: could not add %s: %s", c.URI, err)
			m.mu.Lock()
			delete(m.peers, hostOf(c.URI))
			m.failed[hostOf(c.URI)] = m.now().Add(FailedFor)
			m.mu.Unlock()
			continue
		}
		m.log.Infof("Auto peers: added %s (%s, %.0f ms TCP connect, %s)", c.URI, c.Country, c.ConnectMS, why)
	}
}

// refresh measures candidates from the public catalog, keeps the fastest
// unused ones as standby, and swaps our slowest peer for a clearly faster
// candidate. It uses the network and may take several seconds.
func (m *Manager) refresh(ctx context.Context) {
	candidates, err := m.source(ctx, m.cfg.Country, MaxCount+standbySize)
	links := m.node.GetPeers()
	up := upTransport(links)
	now := m.now()

	m.mu.Lock()
	m.lastFetch = now
	if err != nil {
		m.lastErr = err.Error()
		first := !m.failing
		m.failing = true
		m.mu.Unlock()
		if first {
			m.log.Warnf("Auto peers: could not refresh public peers for %s: %s (keeping current peers and standby list; repeats are logged at debug level)", m.cfg.Country, err)
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
		if !inUse[h] && m.failed[h].IsZero() {
			free = append(free, c)
		}
	}
	sort.SliceStable(free, func(i, j int) bool { return free[i].ConnectMS < free[j].ConnectMS })
	for h, p := range m.peers {
		if ms, ok := byHost[h]; ok {
			p.connectMS = ms
		}
	}

	// One swap per refresh at most: our slowest working peer for a clearly
	// faster free candidate, judged by measured round trips.
	var swapOut *managed
	var swapIn []publicpeers.Suggestion
	if len(free) > 0 {
		var slowest *managed
		for h, p := range m.peers {
			if _, isUp := up[h]; !isUp || p.speed() <= 0 {
				continue
			}
			if slowest == nil || p.speed() > slowest.speed() {
				slowest = p
			}
		}
		best := free[0]
		if slowest != nil && slowest.speed() >
			time.Duration(best.ConnectMS*swapRatio*float64(time.Millisecond))+swapMargin {
			swapOut = slowest
			delete(m.peers, slowest.host)
			h := hostOf(best.URI)
			m.peers[h] = &managed{uri: best.URI, host: h, added: now, connectMS: best.ConnectMS}
			swapIn = append(swapIn, best)
			free = free[1:]
		}
	}
	if len(free) > standbySize {
		free = free[:standbySize]
	}
	m.standby = free
	m.mu.Unlock()

	if swapOut != nil {
		// Add the faster peer first so the node is never short of a link.
		m.addPeers(swapIn, "faster than "+swapOut.uri)
		if u, err := url.Parse(swapOut.uri); err == nil {
			_ = m.node.RemovePeer(u, "")
		}
		m.log.Infof("Auto peers: replaced %s (%s round trip) with a faster peer", swapOut.uri, swapOut.speed().Round(time.Millisecond))
	}
}

// PeerStatus describes one automatic peer.
type PeerStatus struct {
	URI       string  `json:"uri"`
	Up        bool    `json:"up"`
	Added     string  `json:"added"`
	ConnectMS float64 `json:"tcp_connect_ms,omitempty"`
	LatencyMS float64 `json:"latency_ms,omitempty"`
}

// Status is reported by getAutoPeers.
type Status struct {
	Enabled   bool         `json:"enabled"`
	Country   string       `json:"country,omitempty"`
	Count     int          `json:"count,omitempty"`
	Healthy   int          `json:"healthy_peers"`
	Peers     []PeerStatus `json:"auto_peers"`
	Standby   []string     `json:"standby"`
	LastCheck string       `json:"last_check,omitempty"`
	LastError string       `json:"last_error,omitempty"`
}

// Status reports the manager state. A nil manager reports itself disabled.
func (m *Manager) Status() Status {
	if m == nil {
		return Status{Peers: []PeerStatus{}, Standby: []string{}}
	}
	up := upTransport(m.node.GetPeers())
	m.mu.Lock()
	defer m.mu.Unlock()
	s := Status{Enabled: true, Country: m.cfg.Country, Count: m.cfg.Count, Healthy: len(up), Peers: []PeerStatus{}, Standby: []string{}, LastError: m.lastErr}
	if !m.lastCheck.IsZero() {
		s.LastCheck = m.lastCheck.UTC().Format(time.RFC3339)
	}
	for h, p := range m.peers {
		_, isUp := up[h]
		s.Peers = append(s.Peers, PeerStatus{URI: p.uri, Up: isUp, Added: p.added.UTC().Format(time.RFC3339),
			ConnectMS: p.connectMS, LatencyMS: float64(p.latency) / float64(time.Millisecond)})
	}
	sort.Slice(s.Peers, func(i, j int) bool { return s.Peers[i].URI < s.Peers[j].URI })
	for _, c := range m.standby {
		s.Standby = append(s.Standby, c.URI)
	}
	return s
}
