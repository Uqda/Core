package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Uqda/Core/internal/publicpeers"
)

func TestFindOutputAndExitCodes(t *testing.T) {
	for _, test := range []struct {
		name       string
		args       []string
		inJSON     bool
		fetchError bool
		empty      bool
		want       int
	}{
		{"countries", nil, false, false, false, 0},
		{"suggestions", []string{"country=germany"}, false, false, false, 0},
		{"json", []string{"country=germany"}, true, false, false, 0},
		{"none", []string{"country=germany"}, true, false, true, 1},
		{"fetch failed", nil, false, true, false, 1},
		{"invalid before fetch", []string{"password=secret"}, false, false, false, 2},
	} {
		t.Run(test.name, func(t *testing.T) {
			var out, errOut bytes.Buffer
			fetched := false
			fetch := func(context.Context) (publicpeers.Catalog, error) {
				fetched = true
				if test.fetchError {
					return publicpeers.Catalog{}, errors.New("unavailable")
				}
				return publicpeers.Catalog{}, nil
			}
			suggest := func(_ context.Context, _ publicpeers.Catalog, country string, limit int) (publicpeers.Report, error) {
				if limit != 3 {
					t.Fatal(limit)
				}
				report := publicpeers.Report{Source: publicpeers.Source, Updated: time.Now(), Country: country, Countries: []string{"germany"}, Checked: 1, Suggestions: []publicpeers.Suggestion{}, Notice: "No settings changed. TCP reachability only."}
				if !test.empty && country != "" {
					report.Suggestions = append(report.Suggestions, publicpeers.Suggestion{Peer: publicpeers.Peer{URI: "tls://peer.example:443", Uptime: 100}, ConnectMS: 5})
				}
				return report, nil
			}
			code := executeFind(context.Background(), test.args, test.inJSON, &out, &errOut, fetch, suggest)
			if code != test.want || test.want == 2 && fetched || strings.Contains(errOut.String(), "secret") {
				t.Fatal(code, fetched, out.String(), errOut.String())
			}
			if test.inJSON {
				var report publicpeers.Report
				if err := json.Unmarshal(out.Bytes(), &report); err != nil {
					t.Fatal(err, out.String())
				}
			}
			if test.want == 0 && !test.inJSON && !strings.Contains(out.String(), "PUBLIC PEER SUGGESTIONS") {
				t.Fatal(out.String())
			}
		})
	}
}
