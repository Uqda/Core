package autopeer

import (
	"context"
	"crypto/ed25519"
	"errors"
	"net/url"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Uqda/Core/internal/publicpeers"
	"github.com/Uqda/Core/src/core"
)

type fakeNode struct {
	mu      sync.Mutex
	links   map[string]bool // uri -> up
	inbound map[string]bool
	dead    map[string]bool          // added peers that never come up
	latency map[string]time.Duration // live round trip per uri
	added   []string
	removed []string
}

func newFakeNode(static ...string) *fakeNode {
	n := &fakeNode{links: map[string]bool{}, inbound: map[string]bool{}, dead: map[string]bool{}, latency: map[string]time.Duration{}}
	for _, s := range static {
		n.links[s] = true
	}
	return n
}

func (n *fakeNode) AddPeer(u *url.URL, _ string) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.links[u.String()] = !n.dead[u.String()] // the fake connects at once unless dead
	n.added = append(n.added, u.String())
	return nil
}

func (n *fakeNode) RemovePeer(u *url.URL, _ string) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	delete(n.links, u.String())
	n.removed = append(n.removed, u.String())
	return nil
}

func (n *fakeNode) GetPeers() []core.PeerInfo {
	n.mu.Lock()
	defer n.mu.Unlock()
	var out []core.PeerInfo
	for uri, up := range n.links {
		p := core.PeerInfo{URI: uri, Up: up, Inbound: n.inbound[uri], Latency: n.latency[uri]}
		if up {
			p.Key = make(ed25519.PublicKey, ed25519.PublicKeySize)
		}
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].URI < out[j].URI })
	return out
}

func (n *fakeNode) setUp(uri string, up bool) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.links[uri] = up
}

type quietLog struct{}

func (quietLog) Infof(string, ...interface{})  {}
func (quietLog) Warnf(string, ...interface{})  {}
func (quietLog) Debugf(string, ...interface{}) {}

func suggestions(pairs ...interface{}) []publicpeers.Suggestion {
	var out []publicpeers.Suggestion
	for i := 0; i < len(pairs); i += 2 {
		out = append(out, publicpeers.Suggestion{Peer: publicpeers.Peer{URI: pairs[i].(string), Country: "germany"}, ConnectMS: pairs[i+1].(float64)})
	}
	return out
}

type fixedSource struct {
	mu    sync.Mutex
	list  []publicpeers.Suggestion
	err   error
	calls int
}

func (s *fixedSource) get(context.Context, string, int) ([]publicpeers.Suggestion, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls++
	return s.list, s.err
}

// cycle runs one refresh (when asked) and one check, fetching again when
// the check needs candidates, as Run does without its timers.
func (m *Manager) cycle(ctx context.Context, refresh bool) {
	if refresh {
		m.refresh(ctx)
	}
	if m.check() {
		m.refresh(ctx)
		m.check()
	}
}

func newManager(node Node, src *fixedSource, count int) (*Manager, *time.Time) {
	m := New(Config{Country: "germany", Count: count, Interval: time.Hour}, node, src.get, quietLog{})
	now := time.Unix(1_000_000, 0)
	m.now = func() time.Time { return now }
	return m, &now
}

func TestFillsUpToCountWithFastestCandidates(t *testing.T) {
	node := newFakeNode()
	src := &fixedSource{list: suggestions("tls://c.example:443", 90.0, "tls://a.example:443", 20.0, "tls://b.example:443", 40.0)}
	m, _ := newManager(node, src, 2)
	m.cycle(context.Background(), true)
	got := append([]string(nil), node.added...)
	sort.Strings(got)
	if strings.Join(got, ",") != "tls://a.example:443,tls://b.example:443" {
		t.Fatalf("expected the two fastest peers, got %v", got)
	}
	if s := m.Status(); s.Healthy != 2 || len(s.Peers) != 2 {
		t.Fatalf("unexpected status: %+v", s)
	}
}

func TestHealthyStaticPeersMeanNothingIsAdded(t *testing.T) {
	node := newFakeNode("tls://mine-1.example:443", "tls://mine-2.example:443")
	src := &fixedSource{list: suggestions("tls://a.example:443", 1.0)}
	m, _ := newManager(node, src, 2)
	m.cycle(context.Background(), false)
	if len(node.added) != 0 || src.calls != 0 {
		t.Fatalf("nothing should be fetched or added: added=%v calls=%d", node.added, src.calls)
	}
}

func TestStaticPeersAreNeverRemoved(t *testing.T) {
	node := newFakeNode("tls://mine.example:443")
	node.setUp("tls://mine.example:443", false)
	src := &fixedSource{list: suggestions("tls://a.example:443", 10.0)}
	m, now := newManager(node, src, 1)
	for i := 0; i < 10; i++ {
		m.cycle(context.Background(), true)
		*now = now.Add(time.Hour)
	}
	for _, r := range node.removed {
		if r == "tls://mine.example:443" {
			t.Fatal("a configured peer was removed")
		}
	}
	if len(node.added) != 1 || node.added[0] != "tls://a.example:443" {
		t.Fatalf("expected one auto peer to cover the down static peer: %v", node.added)
	}
}

func TestDownAutoPeerIsReplacedAfterGrace(t *testing.T) {
	node := newFakeNode()
	src := &fixedSource{list: suggestions("tls://a.example:443", 10.0, "tls://b.example:443", 20.0)}
	m, now := newManager(node, src, 1)
	m.cycle(context.Background(), true)
	if len(node.added) != 1 || node.added[0] != "tls://a.example:443" {
		t.Fatalf("first pick: %v", node.added)
	}
	node.setUp("tls://a.example:443", false)

	m.cycle(context.Background(), false) // marks it down
	*now = now.Add(DownGrace - time.Second)
	m.cycle(context.Background(), false)
	if len(node.removed) != 0 {
		t.Fatal("peer replaced before the grace period")
	}
	*now = now.Add(2 * time.Second)
	m.cycle(context.Background(), false)
	if len(node.removed) != 1 || node.removed[0] != "tls://a.example:443" {
		t.Fatalf("down peer not removed after grace: %v", node.removed)
	}
	if node.added[len(node.added)-1] != "tls://b.example:443" {
		t.Fatalf("down peer not replaced by the next candidate: %v", node.added)
	}
}

func TestSwapsOnlyForAClearlyFasterPeer(t *testing.T) {
	node := newFakeNode()
	src := &fixedSource{list: suggestions("tls://slow.example:443", 100.0)}
	m, _ := newManager(node, src, 1)
	m.cycle(context.Background(), true)

	// Slightly faster: no churn.
	src.list = suggestions("tls://slow.example:443", 100.0, "tls://bit.example:443", 80.0)
	m.cycle(context.Background(), true)
	if len(node.removed) != 0 {
		t.Fatalf("swapped for a marginal gain: %v", node.removed)
	}
	// Much faster (100 > 20*1.5+30): one swap.
	src.list = suggestions("tls://slow.example:443", 100.0, "tls://fast.example:443", 20.0)
	m.cycle(context.Background(), true)
	if len(node.removed) != 1 || node.removed[0] != "tls://slow.example:443" {
		t.Fatalf("expected a swap of the slow peer: %v", node.removed)
	}
	if node.added[len(node.added)-1] != "tls://fast.example:443" {
		t.Fatalf("expected the fast peer: %v", node.added)
	}
}

func TestCatalogFailureKeepsCurrentPeers(t *testing.T) {
	node := newFakeNode()
	src := &fixedSource{list: suggestions("tls://a.example:443", 10.0)}
	m, _ := newManager(node, src, 1)
	m.cycle(context.Background(), true)
	src.err = errors.New("catalog unreachable")
	m.cycle(context.Background(), true)
	if len(node.removed) != 0 {
		t.Fatal("peers removed because the catalog was unreachable")
	}
	if s := m.Status(); !strings.Contains(s.LastError, "unreachable") || s.Healthy != 1 {
		t.Fatalf("status should report the error and keep the peer: %+v", s)
	}
}

func TestInboundAndLinkLocalLinksDoNotCount(t *testing.T) {
	node := newFakeNode("tls://[fe80::1%25eth0]:1234", "tcp://203.0.113.5:9999")
	node.inbound["tcp://203.0.113.5:9999"] = true
	src := &fixedSource{list: suggestions("tls://a.example:443", 10.0)}
	m, _ := newManager(node, src, 1)
	m.cycle(context.Background(), false)
	if len(node.added) != 1 {
		t.Fatalf("local and inbound links must not satisfy the target: %v", node.added)
	}
}

func TestNilManagerReportsDisabled(t *testing.T) {
	var m *Manager
	if s := m.Status(); s.Enabled || s.Peers == nil {
		t.Fatalf("unexpected status: %+v", s)
	}
}

func TestLostPeerIsReplacedFromStandbyWithoutFetching(t *testing.T) {
	node := newFakeNode()
	src := &fixedSource{list: suggestions("tls://a.example:443", 10.0, "tls://b.example:443", 20.0, "tls://c.example:443", 30.0, "tls://d.example:443", 40.0)}
	m, now := newManager(node, src, 2)
	m.cycle(context.Background(), true)
	if s := m.Status(); len(s.Standby) != 2 || s.Standby[0] != "tls://c.example:443" {
		t.Fatalf("expected c and d on standby: %+v", s.Standby)
	}
	calls := src.calls

	node.setUp("tls://a.example:443", false)
	*now = now.Add(CheckEvery)
	m.check()
	if last := node.added[len(node.added)-1]; last != "tls://c.example:443" {
		t.Fatalf("the fastest standby peer should replace the lost one at once: %v", node.added)
	}
	if src.calls != calls {
		t.Fatal("replacing a lost peer must not wait for a catalog fetch")
	}
	if len(node.removed) != 0 {
		t.Fatal("the lost peer is still retried until DownGrace")
	}
}

func TestPeerThatNeverComesUpIsDroppedAndSkipped(t *testing.T) {
	node := newFakeNode()
	node.dead["tls://a.example:443"] = true
	src := &fixedSource{list: suggestions("tls://a.example:443", 10.0, "tls://b.example:443", 20.0)}
	m, now := newManager(node, src, 1)
	m.cycle(context.Background(), true)
	if len(node.added) != 1 || node.added[0] != "tls://a.example:443" {
		t.Fatalf("first pick: %v", node.added)
	}
	// Still connecting: counted as pending, nothing else is added.
	*now = now.Add(UpWait / 2)
	m.check()
	if len(node.added) != 1 {
		t.Fatalf("a pending peer must not be doubled: %v", node.added)
	}
	*now = now.Add(UpWait)
	m.check()
	if len(node.removed) != 1 || node.removed[0] != "tls://a.example:443" {
		t.Fatalf("a peer that never came up should be removed: %v", node.removed)
	}
	if node.added[len(node.added)-1] != "tls://b.example:443" {
		t.Fatalf("the next candidate should be tried: %v", node.added)
	}
	// A later refresh must not offer the failed peer again.
	m.refresh(context.Background())
	for _, s := range m.Status().Standby {
		if s == "tls://a.example:443" {
			t.Fatal("failed peer offered again before FailedFor")
		}
	}
}

func TestSlowestExtraPeerIsTrimmedOnceStable(t *testing.T) {
	node := newFakeNode()
	src := &fixedSource{list: suggestions("tls://a.example:443", 10.0, "tls://b.example:443", 20.0)}
	m, now := newManager(node, src, 1)
	m.cycle(context.Background(), true)
	node.setUp("tls://a.example:443", false)
	*now = now.Add(CheckEvery)
	m.check() // b replaces a
	// a recovers: two peers up, one wanted. b turns out slower on the live link.
	node.setUp("tls://a.example:443", true)
	node.latency["tls://a.example:443"] = 15 * time.Millisecond
	node.latency["tls://b.example:443"] = 90 * time.Millisecond
	*now = now.Add(CheckEvery)
	m.check()
	if len(node.removed) != 0 {
		t.Fatal("trimmed before the links were stable")
	}
	*now = now.Add(StableFor)
	m.check()
	if len(node.removed) != 1 || node.removed[0] != "tls://b.example:443" {
		t.Fatalf("expected the slower live peer to be trimmed: %v", node.removed)
	}
}

func TestHealthyNodeNeverFetchesBetweenRefreshes(t *testing.T) {
	node := newFakeNode()
	src := &fixedSource{list: suggestions("tls://a.example:443", 10.0)}
	m, now := newManager(node, src, 1)
	m.cycle(context.Background(), true)
	calls := src.calls
	for i := 0; i < 100; i++ {
		*now = now.Add(CheckEvery)
		if m.check() {
			t.Fatal("a healthy node asked for a fetch")
		}
	}
	if src.calls != calls {
		t.Fatal("catalog fetched by local checks")
	}
}

func TestKickNeverBlocks(t *testing.T) {
	m := New(Config{Country: "germany"}, newFakeNode(), (&fixedSource{}).get, quietLog{})
	for i := 0; i < 10; i++ {
		m.Kick()
	}
	var nilManager *Manager
	nilManager.Kick()
}
