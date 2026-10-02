package publicpeers

import (
	"context"
	"fmt"
	"net"
	"net/netip"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const MaxProbes = 8

type Suggestion struct {
	Peer
	ConnectMS float64 `json:"tcp_connect_ms"`
}

type Report struct {
	Source      string       `json:"source"`
	Updated     time.Time    `json:"source_updated_at"`
	CheckedAt   time.Time    `json:"checked_at"`
	Country     string       `json:"country,omitempty"`
	Countries   []string     `json:"countries,omitempty"`
	Checked     int          `json:"checked"`
	Suggestions []Suggestion `json:"suggestions"`
	Notice      string       `json:"notice"`
}

type resolver func(context.Context, string, string) ([]netip.Addr, error)
type dialer func(context.Context, string, string) (net.Conn, error)

func Suggest(ctx context.Context, catalog Catalog, country string, limit int) (Report, error) {
	d := &net.Dialer{Timeout: 1500 * time.Millisecond, KeepAlive: -1}
	return suggest(ctx, catalog, country, limit, net.DefaultResolver.LookupNetIP, d.DialContext)
}

func suggest(ctx context.Context, catalog Catalog, country string, limit int, resolve resolver, dial dialer) (Report, error) {
	report := Report{Source: Source, Updated: catalog.Updated, CheckedAt: time.Now().UTC(), Suggestions: []Suggestion{},
		Notice: "Public TLS peer candidates only. TCP reachability is not a TLS identity, Uqda session or trust check; connect timing is not ping RTT. No daemon, configuration or GroupPassword is changed. Private-group members still require the same group password."}
	if err := catalog.fresh(time.Now()); err != nil {
		return report, err
	}
	if limit < 1 || limit > 5 {
		return report, fmt.Errorf("limit must be 1..5")
	}
	if country == "" {
		report.Countries = catalog.Countries()
		return report, nil
	}
	if !validCountry(country) {
		return report, fmt.Errorf("invalid country; run 'uqda find' for catalog country names")
	}
	report.Country = country
	var candidates []Peer
	for _, peer := range catalog.Peers {
		uri, ok := publicURI(peer.URI)
		if u, _ := url.Parse(uri); ok && u.Scheme == "tls" && peer.Country == country {
			peer.URI = uri
			candidates = append(candidates, peer)
		}
	}
	if len(candidates) == 0 {
		return report, fmt.Errorf("no online TLS candidates for this country; run 'uqda find' for country names")
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].Uptime != candidates[j].Uptime {
			return candidates[i].Uptime > candidates[j].Uptime
		}
		return candidates[i].URI < candidates[j].URI
	})
	seen := map[string]bool{}
	jobs := make(chan Peer, MaxProbes)
	for _, peer := range candidates {
		u, _ := url.Parse(peer.URI)
		hostID := "host:" + strings.ToLower(u.Hostname())
		id := hostID
		if key := u.Query().Get("key"); key != "" {
			id = "key:" + strings.ToLower(key)
		}
		if seen[id] || seen[hostID] {
			continue
		}
		seen[id] = true
		seen[hostID] = true
		jobs <- peer
		report.Checked++
		if report.Checked == MaxProbes {
			break
		}
	}
	close(jobs)
	results := make(chan Suggestion, MaxProbes)
	var workers sync.WaitGroup
	for i := 0; i < 3; i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for peer := range jobs {
				if result, ok := probe(ctx, peer, resolve, dial); ok {
					results <- result
				}
			}
		}()
	}
	workers.Wait()
	close(results)
	for result := range results {
		report.Suggestions = append(report.Suggestions, result)
	}
	sort.Slice(report.Suggestions, func(i, j int) bool {
		if report.Suggestions[i].ConnectMS != report.Suggestions[j].ConnectMS {
			return report.Suggestions[i].ConnectMS < report.Suggestions[j].ConnectMS
		}
		return report.Suggestions[i].URI < report.Suggestions[j].URI
	})
	if len(report.Suggestions) > limit {
		report.Suggestions = report.Suggestions[:limit]
	}
	return report, nil
}

func probe(parent context.Context, peer Peer, resolve resolver, dial dialer) (Suggestion, bool) {
	ctx, cancel := context.WithTimeout(parent, 3*time.Second)
	defer cancel()
	uri, valid := publicURI(peer.URI)
	if !valid {
		return Suggestion{}, false
	}
	u, _ := url.Parse(uri)
	var addresses []netip.Addr
	if ip, err := netip.ParseAddr(u.Hostname()); err == nil {
		addresses = []netip.Addr{ip}
	} else {
		var err error
		addresses, err = resolve(ctx, "ip", u.Hostname())
		if err != nil {
			return Suggestion{}, false
		}
	}
	if len(addresses) == 0 || len(addresses) > 16 {
		return Suggestion{}, false
	}
	// Reject mixed public/private DNS answers, then dial only resolved literal
	// public addresses. Never resolve the hostname a second time while dialing.
	for _, ip := range addresses {
		if !publicIP(ip) {
			return Suggestion{}, false
		}
	}
	sort.SliceStable(addresses, func(i, j int) bool { return addresses[i].Is4() && !addresses[j].Is4() })
	for i, ip := range addresses {
		if i == 2 || ctx.Err() != nil {
			break
		}
		start := time.Now()
		conn, err := dial(ctx, "tcp", net.JoinHostPort(ip.Unmap().String(), u.Port()))
		if err == nil {
			elapsed := float64(time.Since(start)) / float64(time.Millisecond)
			_ = conn.Close()
			return Suggestion{Peer: peer, ConnectMS: elapsed}, true
		}
	}
	return Suggestion{}, false
}

// ParseOptions intentionally accepts only small, explicit country/limit inputs.
// There is no custom URL, arbitrary target, auto-connect or configuration option.
func ParseOptions(args []string) (string, int, error) {
	country, limit := "", 3
	seen := map[string]bool{}
	for _, arg := range args {
		key, value, ok := strings.Cut(arg, "=")
		if !ok || seen[key] {
			return "", 0, fmt.Errorf("use find [country=NAME] [limit=1..5]")
		}
		seen[key] = true
		switch key {
		case "country":
			country = strings.ToLower(value)
			if !validCountry(country) {
				return "", 0, fmt.Errorf("use a catalog country name, for example country=germany")
			}
		case "limit":
			var err error
			limit, err = strconv.Atoi(value)
			if err != nil || limit < 1 || limit > 5 {
				return "", 0, fmt.Errorf("limit must be 1..5")
			}
		default:
			return "", 0, fmt.Errorf("find supports only country=NAME and limit=1..5")
		}
	}
	return country, limit, nil
}
