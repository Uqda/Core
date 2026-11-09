package core

import (
	"encoding/binary"
	"net"
	"net/url"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Uqda/Core/src/config"
)

// cutProxy forwards TCP between a node and its peer. cut(true) closes both
// sides, like a peer shutting down; cut(false) silently drops all bytes
// while keeping the sockets open, like a router or Wi-Fi link that died.
type cutProxy struct {
	ln      net.Listener
	target  string
	dropped atomic.Bool
	mu      sync.Mutex
	conns   []net.Conn
}

func newCutProxy(t *testing.T, target string) *cutProxy {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	p := &cutProxy{ln: ln, target: target}
	go p.serve()
	t.Cleanup(func() { p.cut(true) })
	return p
}

func (p *cutProxy) serve() {
	for {
		c, err := p.ln.Accept()
		if err != nil {
			return
		}
		u, err := net.Dial("tcp", p.target)
		if err != nil {
			_ = c.Close()
			continue
		}
		p.mu.Lock()
		p.conns = append(p.conns, c, u)
		p.mu.Unlock()
		go p.pipe(c, u)
		go p.pipe(u, c)
	}
}

func (p *cutProxy) pipe(dst, src net.Conn) {
	buf := make([]byte, 32*1024)
	for {
		n, err := src.Read(buf)
		if n > 0 && !p.dropped.Load() {
			if _, werr := dst.Write(buf[:n]); werr != nil {
				return
			}
		}
		if err != nil {
			_ = dst.Close() // EOF or error: pass the close on
			return
		}
	}
}

func (p *cutProxy) url() *url.URL {
	return &url.URL{Scheme: "tcp", Host: p.ln.Addr().String()}
}

func (p *cutProxy) cut(closeConns bool) {
	p.dropped.Store(true)
	if !closeConns {
		return
	}
	_ = p.ln.Close()
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, c := range p.conns {
		_ = c.Close()
	}
}

func newQuietNode(t *testing.T) *Core {
	t.Helper()
	cfg := config.GenerateConfig()
	if err := cfg.GenerateSelfSignedCertificate(); err != nil {
		t.Fatal(err)
	}
	n, err := New(cfg.Certificate, GetLoggerWithPrefix("", false))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(n.Stop)
	return n
}

// TestFailoverAroundLostRelay sends a steady stream from A to C while A has
// two independent paths (via B and via D), cuts the link carrying the
// traffic, and measures how long delivery stops and how much is lost.
func TestFailoverAroundLostRelay(t *testing.T) {
	if testing.Short() {
		t.Skip("failover measurement takes about 30 seconds")
	}
	for _, tc := range []struct {
		name     string
		closeIt  bool
		maxGap   time.Duration
		maxShare float64 // maximum share of packets lost after the cut
	}{
		// The peer's connection closes: detected at once. Usually traffic
		// moves within milliseconds; when the lost link was this node's
		// parent in the tree, the tree re-forms first (about 3-4 s).
		{"peer closes", true, 6 * time.Second, 0.5},
		// Packets silently vanish: detected by the keepalive timeout (3 s).
		{"link goes silent", false, 9 * time.Second, 0.7},
	} {
		t.Run(tc.name, func(t *testing.T) { runFailover(t, tc.closeIt, tc.maxGap, tc.maxShare) })
	}
}

func runFailover(t *testing.T, closeIt bool, maxGap time.Duration, maxShare float64) {
	a, b, c, d := newQuietNode(t), newQuietNode(t), newQuietNode(t), newQuietNode(t)
	cAddr := listenTCP(t, c)
	for _, relay := range []*Core{b, d} {
		if err := relay.AddPeer(&url.URL{Scheme: "tcp", Host: cAddr}, ""); err != nil {
			t.Fatal(err)
		}
	}
	viaB := newCutProxy(t, listenTCP(t, b))
	viaD := newCutProxy(t, listenTCP(t, d))
	for _, p := range []*cutProxy{viaB, viaD} {
		if err := a.AddPeer(p.url(), ""); err != nil {
			t.Fatal(err)
		}
	}
	// Every node has both of its links up. (A node's tree view only holds
	// the entries it needs, so the tree size is not a readiness signal.)
	if !waitForCondition(20*time.Second, func() bool {
		for _, n := range []*Core{a, b, c, d} {
			up := 0
			for _, p := range n.GetPeers() {
				if p.Up {
					up++
				}
			}
			if up < 2 {
				return false
			}
		}
		return true
	}) {
		for name, n := range map[string]*Core{"A": a, "B": b, "C": c, "D": d} {
			for _, p := range n.GetPeers() {
				t.Logf("node %s peer %s up=%v inbound=%v err=%v", name, p.URI, p.Up, p.Inbound, p.LastError)
			}
		}
		t.Fatal("the four nodes did not connect")
	}

	// Every node reads, as the TUN adapter does in the daemon: session
	// handshakes are completed on the read path, so a node that never reads
	// cannot finish setting up a session.
	for _, n := range []*Core{a, b, d} {
		go func(n *Core) {
			buf := make([]byte, 2048)
			for {
				if _, _, err := n.ReadFrom(buf); err != nil {
					return
				}
			}
		}(n)
	}

	// Receiver: record when each sequence number arrives at C.
	var mu sync.Mutex
	arrived := map[uint64]time.Time{}
	go func() {
		buf := make([]byte, 2048)
		for {
			n, _, err := c.ReadFrom(buf)
			if err != nil {
				return
			}
			if n >= 48 {
				mu.Lock()
				arrived[binary.BigEndian.Uint64(buf[40:48])] = time.Now()
				mu.Unlock()
			}
		}
	}()

	// Sender: one packet every 10 ms from A to C.
	const every = 10 * time.Millisecond
	sentAt := map[uint64]time.Time{}
	var seq uint64
	send := func(d time.Duration) {
		end := time.Now().Add(d)
		for time.Now().Before(end) {
			msg := make([]byte, 100)
			msg[0] = 0x60
			copy(msg[8:24], a.Address())
			copy(msg[24:40], c.Address())
			binary.BigEndian.PutUint64(msg[40:48], seq)
			mu.Lock()
			sentAt[seq] = time.Now()
			mu.Unlock()
			_, _ = a.WriteTo(msg, c.LocalAddr())
			seq++
			time.Sleep(every)
		}
	}

	// Send until the session is established and packets flow end to end.
	flowing := func() bool {
		send(200 * time.Millisecond)
		mu.Lock()
		defer mu.Unlock()
		return len(arrived) >= 10
	}
	if !waitForCondition(20*time.Second, flowing) {
		t.Fatal("no traffic reached C before the cut")
	}

	// Warm up, and find which relay carries the traffic.
	txTo := func(relay *Core) uint64 {
		for _, p := range a.GetPeers() {
			if p.Up && p.Key.Equal(relay.PublicKey()) {
				return p.TXBytes
			}
		}
		return 0
	}
	b0, d0 := txTo(b), txTo(d)
	send(2 * time.Second)
	active, name := viaB, "B"
	if txTo(d)-d0 > txTo(b)-b0 {
		active, name = viaD, "D"
	}
	mu.Lock()
	before := len(arrived)
	mu.Unlock()
	if before == 0 {
		t.Fatal("no traffic arrived before the cut")
	}

	cutSeq := seq
	cutAt := time.Now()
	active.cut(closeIt)
	send(12 * time.Second)
	time.Sleep(time.Second)

	mu.Lock()
	defer mu.Unlock()
	lastBefore, firstAfter := time.Time{}, time.Time{}
	for s, at := range arrived {
		if s < cutSeq && at.After(lastBefore) {
			lastBefore = at
		}
		if s >= cutSeq && (firstAfter.IsZero() || at.Before(firstAfter)) {
			firstAfter = at
		}
	}
	if firstAfter.IsZero() {
		t.Fatalf("traffic never resumed after cutting the path via %s", name)
	}
	if lastBefore.IsZero() || lastBefore.Before(cutAt.Add(-time.Second)) {
		lastBefore = cutAt
	}
	gap := firstAfter.Sub(lastBefore)
	sentAfter := seq - cutSeq
	var got uint64
	tailSent, tailGot := 0, 0
	tail := cutAt.Add(8 * time.Second)
	for s := cutSeq; s < seq; s++ {
		_, ok := arrived[s]
		if ok {
			got++
		}
		if sentAt[s].After(tail) {
			tailSent++
			if ok {
				tailGot++
			}
		}
	}
	lost := sentAfter - got
	t.Logf("cut via %s: delivery stopped for %s; %d of %d packets lost (%.1f%%); after recovery %d of %d delivered",
		name, gap.Round(time.Millisecond), lost, sentAfter, 100*float64(lost)/float64(sentAfter), tailGot, tailSent)
	if gap > maxGap {
		t.Fatalf("delivery stopped for %s, more than %s", gap, maxGap)
	}
	if share := float64(lost) / float64(sentAfter); share > maxShare {
		t.Fatalf("lost %.0f%% of packets after the cut, more than %.0f%%", 100*share, 100*maxShare)
	}
	if tailSent > 0 && float64(tailGot) < 0.98*float64(tailSent) {
		t.Fatalf("traffic did not fully recover: %d of %d delivered", tailGot, tailSent)
	}
}
