package publicpeers

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/netip"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestOptions(t *testing.T) {
	country, limit, err := ParseOptions([]string{"country=GERMANY", "limit=2"})
	if err != nil || country != "germany" || limit != 2 {
		t.Fatal(country, limit, err)
	}
	for _, args := range [][]string{{"country="}, {"country=../secret"}, {"limit=0"}, {"limit=6"}, {"limit=abc"}, {"country=germany", "country=austria"}, {"url=https://secret.example"}, {"password=secret"}, {"germany"}} {
		if _, _, err := ParseOptions(args); err == nil || strings.Contains(err.Error(), "secret") {
			t.Fatal("unsafe usage", args, err)
		}
	}
}

func TestRejectDNSBeforeDial(t *testing.T) {
	peer := Peer{URI: "tls://peer.example:443"}
	for _, raw := range [][]string{{"127.0.0.1"}, {"8.8.8.8", "10.0.0.1"}, {"169.254.169.254"}, {"::ffff:127.0.0.1"}, {}} {
		resolve := func(context.Context, string, string) ([]netip.Addr, error) {
			var result []netip.Addr
			for _, ip := range raw {
				result = append(result, netip.MustParseAddr(ip))
			}
			return result, nil
		}
		dial := func(context.Context, string, string) (net.Conn, error) {
			t.Error("unsafe address dialed")
			return nil, errors.New("unsafe")
		}
		if _, ok := probe(context.Background(), peer, resolve, dial); ok {
			t.Fatal("unsafe peer recommended")
		}
	}
}

func TestProbePinsResolvedIPAndClosesConnection(t *testing.T) {
	local, remote := net.Pipe()
	defer remote.Close()
	resolve := func(context.Context, string, string) ([]netip.Addr, error) {
		return []netip.Addr{netip.MustParseAddr("8.8.8.8")}, nil
	}
	dial := func(ctx context.Context, network, address string) (net.Conn, error) {
		if network != "tcp" || address != "8.8.8.8:443" {
			t.Fatal("not pinned", network, address)
		}
		if deadline, ok := ctx.Deadline(); !ok || time.Until(deadline) > 3*time.Second {
			t.Fatal("unbounded dial")
		}
		return local, nil
	}
	if result, ok := probe(context.Background(), Peer{URI: "tls://peer.example:443"}, resolve, dial); !ok || result.ConnectMS < 0 {
		t.Fatal(result, ok)
	}
	if _, err := remote.Write([]byte("x")); err == nil {
		t.Fatal("connection not closed")
	}
}

func TestSuggestionBudgetDeduplicationAndReadOnlyCountries(t *testing.T) {
	catalog := Catalog{Updated: time.Now()}
	for i := 0; i < 20; i++ {
		catalog.Peers = append(catalog.Peers, Peer{URI: fmt.Sprintf("tls://peer%d.example:443", i), Country: "germany", Uptime: 100 - i})
	}
	catalog.Peers = append(catalog.Peers, Peer{URI: "tls://peer0.example:444?key=" + strings.Repeat("a", 64), Country: "germany", Uptime: 100})
	var calls, active, peak atomic.Int32
	resolve := func(context.Context, string, string) ([]netip.Addr, error) {
		return []netip.Addr{netip.MustParseAddr("8.8.8.8")}, nil
	}
	dial := func(context.Context, string, string) (net.Conn, error) {
		calls.Add(1)
		current := active.Add(1)
		defer active.Add(-1)
		for old := peak.Load(); current > old; old = peak.Load() {
			if peak.CompareAndSwap(old, current) {
				break
			}
		}
		local, remote := net.Pipe()
		remote.Close()
		return local, nil
	}
	report, err := suggest(context.Background(), catalog, "", 3, resolve, dial)
	if err != nil || calls.Load() != 0 || len(report.Countries) != 1 {
		t.Fatal(report, err)
	}
	report, err = suggest(context.Background(), catalog, "germany", 3, resolve, dial)
	if err != nil || calls.Load() != MaxProbes || report.Checked != MaxProbes || len(report.Suggestions) != 3 || peak.Load() > 3 {
		t.Fatal(report, calls.Load(), peak.Load(), err)
	}
	for i := 1; i < len(report.Suggestions); i++ {
		if report.Suggestions[i].ConnectMS < report.Suggestions[i-1].ConnectMS {
			t.Fatal("not ranked")
		}
	}
	before := calls.Load()
	catalog.Updated = time.Now().Add(-7 * time.Hour)
	if _, err := suggest(context.Background(), catalog, "germany", 3, resolve, dial); err == nil || calls.Load() != before {
		t.Fatal("stale source probed")
	}
}

func TestCancelledAndFailedProbes(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	resolve := func(context.Context, string, string) ([]netip.Addr, error) {
		return []netip.Addr{netip.MustParseAddr("8.8.8.8")}, nil
	}
	dial := func(context.Context, string, string) (net.Conn, error) {
		t.Error("cancelled probe dialed")
		return nil, errors.New("failed")
	}
	if _, ok := probe(ctx, Peer{URI: "tls://peer.example:443"}, resolve, dial); ok {
		t.Fatal("cancelled probe succeeded")
	}
	var calls int
	dial = func(context.Context, string, string) (net.Conn, error) {
		calls++
		return nil, errors.New("secret transport details")
	}
	catalog := Catalog{Updated: time.Now(), Peers: []Peer{{URI: "tls://peer.example:443", Country: "germany"}}}
	report, err := suggest(context.Background(), catalog, "germany", 3, resolve, dial)
	if err != nil || calls != 1 || len(report.Suggestions) != 0 || strings.Contains(report.Notice, "secret transport") {
		t.Fatal(report, calls, err)
	}
}

func TestPinnedKeyDedupAndAddressAttemptBudget(t *testing.T) {
	key := strings.Repeat("a", 64)
	catalog := Catalog{Updated: time.Now(), Peers: []Peer{
		{URI: "tls://a.example:443?key=" + key, Country: "germany", Uptime: 100},
		{URI: "tls://b.example:443?key=" + key, Country: "germany", Uptime: 99},
	}}
	resolve := func(context.Context, string, string) ([]netip.Addr, error) {
		return []netip.Addr{netip.MustParseAddr("8.8.8.8"), netip.MustParseAddr("8.8.4.4"), netip.MustParseAddr("1.1.1.1")}, nil
	}
	var calls atomic.Int32
	dial := func(context.Context, string, string) (net.Conn, error) {
		calls.Add(1)
		return nil, errors.New("unreachable")
	}
	report, err := suggest(context.Background(), catalog, "germany", 3, resolve, dial)
	if err != nil || report.Checked != 1 || calls.Load() != 2 || len(report.Suggestions) != 0 {
		t.Fatal(report, calls.Load(), err)
	}
}
