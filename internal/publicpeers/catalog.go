// Package publicpeers provides bounded, read-only public peer suggestions.
package publicpeers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/netip"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"
)

const Source = "https://publicpeers.neilalexander.dev/"
const maxBody = 1024 * 1024

type Peer struct {
	URI     string `json:"uri"`
	Country string `json:"country"`
	Uptime  int    `json:"site_uptime_percent"`
}

type Catalog struct {
	Updated time.Time `json:"source_updated_at"`
	Peers   []Peer    `json:"-"`
}

// Fetch contacts only the fixed HTTPS catalog; no user-supplied source or
// cross-origin redirect can redirect the request into an internal network.
func Fetch(ctx context.Context) (Catalog, error) {
	client := &http.Client{Timeout: 10 * time.Second, CheckRedirect: catalogRedirect}
	return fetch(ctx, client, Source)
}

func catalogRedirect(req *http.Request, via []*http.Request) error {
	if len(via) >= 3 || req.URL.Scheme != "https" || req.URL.Host != "publicpeers.neilalexander.dev" || req.URL.User != nil {
		return fmt.Errorf("catalog redirect refused")
	}
	return nil
}

func fetch(ctx context.Context, client *http.Client, source string) (Catalog, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, source, nil)
	if err != nil {
		return Catalog{}, fmt.Errorf("invalid catalog request")
	}
	req.Header.Set("User-Agent", "Uqda/public-peer-discovery")
	req.Header.Set("Accept", "text/html")
	response, err := client.Do(req)
	if err != nil {
		return Catalog{}, fmt.Errorf("cannot fetch public peer catalog; check Internet access")
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return Catalog{}, fmt.Errorf("catalog returned HTTP %d", response.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, maxBody+1))
	if err != nil || len(body) > maxBody {
		return Catalog{}, fmt.Errorf("catalog response is unreadable or exceeds 1 MiB")
	}
	catalog, err := parseCatalog(body)
	if err != nil {
		return Catalog{}, err
	}
	if err := catalog.fresh(time.Now()); err != nil {
		return Catalog{}, err
	}
	return catalog, nil
}

func (c Catalog) fresh(now time.Time) error {
	if c.Updated.IsZero() || c.Updated.After(now.Add(5*time.Minute)) || now.Sub(c.Updated) > 6*time.Hour {
		return fmt.Errorf("catalog timestamp is missing, stale or in the future; no peers will be probed")
	}
	return nil
}

func parseCatalog(body []byte) (Catalog, error) {
	if len(body) > maxBody {
		return Catalog{}, fmt.Errorf("catalog exceeds 1 MiB")
	}
	var catalog Catalog
	z := html.NewTokenizer(strings.NewReader(string(body)))
	var country, field, stamp string
	var row map[string]string
	inStamp := false
	for {
		switch z.Next() {
		case html.ErrorToken:
			if z.Err() != io.EOF {
				return Catalog{}, fmt.Errorf("invalid catalog HTML")
			}
			if catalog.Updated.IsZero() || len(catalog.Peers) == 0 {
				return Catalog{}, fmt.Errorf("catalog format is unsupported or contains no usable public peers")
			}
			return catalog, nil
		case html.StartTagToken:
			t := z.Token()
			if t.Data == "strong" && catalog.Updated.IsZero() {
				inStamp, stamp = true, ""
			}
			if t.Data == "tr" {
				row, field = map[string]string{}, ""
			}
			if t.Data == "td" || t.Data == "th" {
				field = ""
				for _, a := range t.Attr {
					if a.Key == "id" {
						field = a.Val
					}
				}
				if row != nil && field != "" {
					row[field] = ""
				}
			}
		case html.TextToken:
			text := string(z.Text())
			if inStamp {
				if len(stamp)+len(text) >= 80 {
					return Catalog{}, fmt.Errorf("catalog timestamp exceeds safe length")
				}
				stamp += text
			}
			if row != nil && field != "" && len(row[field])+len(text) <= 1024 {
				row[field] += text
			} else if row != nil && field != "" {
				return Catalog{}, fmt.Errorf("catalog field exceeds safe length")
			}
		case html.EndTagToken:
			switch z.Token().Data {
			case "strong":
				if inStamp {
					if ts, err := time.Parse("02 Jan 06 15:04 MST", strings.TrimSpace(stamp)); err == nil && strings.HasSuffix(strings.TrimSpace(stamp), " UTC") {
						catalog.Updated = ts.UTC()
					}
					inStamp = false
				}
			case "td", "th":
				field = ""
			case "tr":
				if name, present := row["country"]; present {
					country = ""
					if name = strings.TrimSpace(name); validCountry(name) {
						country = name
					}
				}
				status := strings.Fields(row["status"])
				uri, ok := publicURI(strings.TrimSpace(row["address"]))
				uptimeText := strings.TrimSpace(row["reliability"])
				uptime, err := strconv.Atoi(strings.TrimSuffix(uptimeText, "%"))
				if country != "" && ok && len(status) > 0 && status[0] == "online" && err == nil && strings.HasSuffix(uptimeText, "%") && uptime >= 0 && uptime <= 100 {
					catalog.Peers = append(catalog.Peers, Peer{URI: uri, Country: country, Uptime: uptime})
				}
				row, field = nil, ""
			}
		}
	}
}

func validCountry(s string) bool {
	if s == "" || len(s) > 64 || s[0] < 'a' || s[0] > 'z' {
		return false
	}
	for _, c := range s {
		if (c < 'a' || c > 'z') && c != '-' {
			return false
		}
	}
	return true
}

func (c Catalog) Countries() []string {
	seen := map[string]bool{}
	for _, p := range c.Peers {
		uri, ok := publicURI(p.URI)
		if u, _ := url.Parse(uri); ok && u.Scheme == "tls" && validCountry(p.Country) {
			seen[p.Country] = true
		}
	}
	names := make([]string, 0, len(seen))
	for name := range seen {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func publicURI(raw string) (string, bool) {
	u, err := url.Parse(raw)
	if err != nil || len(raw) > 512 || (u.Scheme != "tls" && u.Scheme != "tcp") || u.User != nil || u.Path != "" || u.Fragment != "" || u.Opaque != "" {
		return "", false
	}
	port, err := strconv.Atoi(u.Port())
	if err != nil || port < 1 || port > 65535 {
		return "", false
	}
	host := u.Hostname()
	if ip, err := netip.ParseAddr(host); err == nil {
		if !publicIP(ip) {
			return "", false
		}
	} else if !validDomain(host) {
		return "", false
	}
	q, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		return "", false
	}
	for key, values := range q {
		if len(values) != 1 {
			return "", false
		}
		switch key {
		case "key":
			if len(values[0]) != 64 || strings.IndexFunc(values[0], func(c rune) bool { return !strings.ContainsRune("0123456789abcdefABCDEF", c) }) >= 0 {
				return "", false
			}
		case "sni":
			if !validDomain(values[0]) {
				return "", false
			}
		default:
			return "", false
		}
	}
	u.RawQuery = q.Encode()
	return u.String(), true
}

func validDomain(s string) bool {
	if len(s) > 253 || !strings.Contains(s, ".") {
		return false
	}
	for _, label := range strings.Split(strings.TrimSuffix(s, "."), ".") {
		if label == "" || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}
		for _, c := range label {
			if c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '-' {
				continue
			}
			return false
		}
	}
	return true
}

var excluded = []netip.Prefix{
	netip.MustParsePrefix("0.0.0.0/8"), netip.MustParsePrefix("100.64.0.0/10"),
	netip.MustParsePrefix("192.0.0.0/24"), netip.MustParsePrefix("192.0.2.0/24"),
	netip.MustParsePrefix("192.88.99.0/24"), netip.MustParsePrefix("198.18.0.0/15"),
	netip.MustParsePrefix("198.51.100.0/24"), netip.MustParsePrefix("203.0.113.0/24"),
	netip.MustParsePrefix("240.0.0.0/4"), netip.MustParsePrefix("2001::/23"),
	netip.MustParsePrefix("2001:db8::/32"), netip.MustParsePrefix("2002::/16"),
}

func publicIP(ip netip.Addr) bool {
	if ip.Zone() != "" {
		return false
	}
	ip = ip.Unmap()
	if !ip.IsGlobalUnicast() || ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast() {
		return false
	}
	if ip.Is6() && !netip.MustParsePrefix("2000::/3").Contains(ip) {
		return false
	}
	for _, prefix := range excluded {
		if prefix.Contains(ip) {
			return false
		}
	}
	return true
}
