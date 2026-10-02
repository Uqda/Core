package publicpeers

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"
)

func fixture(stamp time.Time, rows string) string {
	return "<strong>" + stamp.UTC().Format("02 Jan 06 15:04 MST") + "</strong><table>" + rows + "</table>"
}

func heading(country string) string {
	return "<thead><tr><th id='country'>" + country + "</th></tr></thead>"
}

func row(uri, status, uptime string) string {
	return fmt.Sprintf("<tbody><tr><td id='address'>%s</td><td id='status'>%s</td><td id='reliability'>%s</td></tr></tbody>", uri, status, uptime)
}

func TestCatalogFiltersAndNormalizes(t *testing.T) {
	stamp := time.Now().UTC().Truncate(time.Minute)
	body := fixture(stamp, heading("germany")+
		row("tls://peer.example:1234?sni=peer.example&amp;key="+strings.Repeat("a", 64), "online 1 week+", "100%")+
		row("tls://offline.example:1234", "offline", "99%")+
		row("tls://127.0.0.1:1234", "online", "100%")+
		row("tls://secret.example:1234?password=secret", "online", "100%")+
		row("tls://bad.example:1234", "online", "101%")+
		heading("austria")+row("tls://other.example:1234", "online", "98%")+
		heading("invalid country")+row("tls://misclassified.example:1234", "online", "100%")+
		heading("")+row("tls://empty-country.example:1234", "online", "100%"))
	catalog, err := parseCatalog([]byte(body))
	if err != nil || len(catalog.Peers) != 2 || !catalog.Updated.Equal(stamp) {
		t.Fatalf("catalog=%+v error=%v", catalog, err)
	}
	if !reflect.DeepEqual(catalog.Countries(), []string{"austria", "germany"}) {
		t.Fatal(catalog.Countries())
	}
	if !strings.Contains(catalog.Peers[0].URI, "key="+strings.Repeat("a", 64)+"&sni=") {
		t.Fatal("HTML entity or canonical query lost", catalog.Peers[0].URI)
	}
}

func TestUnsafeURIsAndAddresses(t *testing.T) {
	for _, raw := range []string{
		"tls://localhost:1", "tls://127.0.0.1:1", "tls://10.1.2.3:1", "tls://169.254.169.254:80",
		"tls://100.64.0.1:1", "tls://192.0.2.1:1", "tls://224.0.0.1:1", "tls://198.18.0.1:1",
		"tls://[::1]:1", "tls://[fc00::1]:1", "tls://[fe80::1%25en0]:1", "tls://[::ffff:127.0.0.1]:1",
		"tls://[2001:db8::1]:1", "tls://[2002:7f00:1::]:1", "tls://[64:ff9b::7f00:1]:1",
		"http://peer.example:1", "quic://peer.example:1", "tls://user:password@peer.example:1",
		"tls://peer.example:0", "tls://peer.example:65536", "tls://peer.example", "tls://peer.example:1/path",
		"tls://peer.example:1#secret", "tls://peer.example:1?password=secret", "tls://peer.example:1?sni=x%0aevil",
		"tls://peer.example:1?key=abc", "tls://peer.example:1?sni=a.example&sni=b.example",
		"tls://peer.example:1?unknown=x", "tls://peer.example:1?key=%zz", "tls://\x1b.example:1",
	} {
		t.Run(raw, func(t *testing.T) {
			if _, ok := publicURI(raw); ok {
				t.Fatal("unsafe URI accepted")
			}
		})
	}
	for _, raw := range []string{"tls://peer.example:443", "tls://8.8.8.8:443", "tls://[2606:4700:4700::1111]:443"} {
		if _, ok := publicURI(raw); !ok {
			t.Fatal("public URI rejected", raw)
		}
	}
	if publicIP(netip.Addr{}) {
		t.Fatal("invalid IP accepted")
	}
}

func TestCatalogBoundsAndFreshness(t *testing.T) {
	for _, body := range []string{"", "<strong>not a timestamp</strong>", strings.Repeat("x", maxBody+1),
		fixture(time.Now(), heading("germany")+row(strings.Repeat("x", 1025), "online", "100%"))} {
		if _, err := parseCatalog([]byte(body)); err == nil {
			t.Fatal("invalid catalog accepted")
		}
	}
	now := time.Now()
	for _, stamp := range []time.Time{{}, now.Add(-7 * time.Hour), now.Add(6 * time.Minute)} {
		if (Catalog{Updated: stamp}).fresh(now) == nil {
			t.Fatal("unsafe timestamp accepted", stamp)
		}
	}
	if err := (Catalog{Updated: now.Add(-time.Hour)}).fresh(now); err != nil {
		t.Fatal(err)
	}
}

func TestFetchAndRedirectLimits(t *testing.T) {
	for _, test := range []struct {
		name      string
		status    int
		body      string
		wantError bool
	}{
		{"valid", 200, fixture(time.Now(), heading("germany")+row("tls://peer.example:443", "online", "100%")), false},
		{"stale", 200, fixture(time.Now().Add(-7*time.Hour), heading("germany")+row("tls://peer.example:443", "online", "100%")), true},
		{"oversize", 200, strings.Repeat("x", maxBody+1), true},
		{"status", 503, "unavailable", true},
	} {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "GET" || r.Header.Get("User-Agent") != "Uqda/public-peer-discovery" {
					t.Error("unexpected request")
				}
				w.WriteHeader(test.status)
				fmt.Fprint(w, test.body)
			}))
			defer server.Close()
			_, err := fetch(context.Background(), server.Client(), server.URL)
			if (err != nil) != test.wantError {
				t.Fatal(err)
			}
		})
	}
	for _, raw := range []string{"http://publicpeers.neilalexander.dev/", "https://evil.example/", "https://publicpeers.neilalexander.dev:8443/", "https://user@publicpeers.neilalexander.dev/"} {
		u, _ := url.Parse(raw)
		if catalogRedirect(&http.Request{URL: u}, nil) == nil {
			t.Fatal("unsafe redirect allowed", raw)
		}
	}
	u, _ := url.Parse(Source)
	if catalogRedirect(&http.Request{URL: u}, nil) != nil {
		t.Fatal("same-origin redirect refused")
	}
	if catalogRedirect(&http.Request{URL: u}, make([]*http.Request, 3)) == nil {
		t.Fatal("redirect loop allowed")
	}
}

func FuzzCatalog(f *testing.F) {
	f.Add([]byte(fixture(time.Now(), heading("germany")+row("tls://peer.example:443", "online", "100%"))))
	f.Add([]byte("<table><tr><td id='address'>bad</td></tr></table>"))
	f.Fuzz(func(t *testing.T, data []byte) {
		catalog, err := parseCatalog(data)
		if err != nil {
			return
		}
		for _, peer := range catalog.Peers {
			if _, ok := publicURI(peer.URI); !ok || !validCountry(peer.Country) {
				t.Fatal("unsafe parsed peer")
			}
		}
	})
}
