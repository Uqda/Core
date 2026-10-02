package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/Uqda/Core/internal/cli"
	"github.com/Uqda/Core/internal/publicpeers"
	"github.com/Uqda/Core/src/version"
)

func runFind(args []string, inJSON bool) int {
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()
	return executeFind(ctx, args, inJSON, os.Stdout, os.Stderr, publicpeers.Fetch, publicpeers.Suggest)
}

func executeFind(ctx context.Context, args []string, inJSON bool, out, errOut io.Writer,
	fetch func(context.Context) (publicpeers.Catalog, error),
	suggest func(context.Context, publicpeers.Catalog, string, int) (publicpeers.Report, error)) int {
	country, limit, err := publicpeers.ParseOptions(args)
	if err != nil {
		fmt.Fprintln(errOut, "Uqda:", err, "Run 'uqda help find'.")
		return 2
	}
	catalog, err := fetch(ctx)
	if err != nil {
		fmt.Fprintln(errOut, "Uqda:", err)
		return 1
	}
	report, err := suggest(ctx, catalog, country, limit)
	if err != nil {
		fmt.Fprintln(errOut, "Uqda:", err)
		return 1
	}
	if inJSON {
		encoder := json.NewEncoder(out)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(report); err != nil {
			fmt.Fprintln(errOut, "Uqda: cannot write peer suggestions")
			return 1
		}
	} else {
		renderFind(out, report)
	}
	if country != "" && len(report.Suggestions) == 0 {
		return 1
	}
	return 0
}

func renderFind(w io.Writer, report publicpeers.Report) {
	cli.Header(w, "PUBLIC PEER SUGGESTIONS", version.DisplayName())
	cli.Field(w, "Source", report.Source)
	cli.Field(w, "Site updated", report.Updated.UTC().Format("2006-01-02 15:04 UTC"))
	if report.Country == "" {
		fmt.Fprintln(w, "\n  Available countries (online TLS candidates):")
		for _, country := range report.Countries {
			fmt.Fprintln(w, "    "+country)
		}
		fmt.Fprintln(w, "\n  Next: uqda find country=germany limit=3")
		cli.Text(w, "  ", "Choose a country near you. No location lookup or peer connections were made.")
	} else {
		cli.Field(w, "Country", report.Country)
		cli.Field(w, "Candidates", fmt.Sprintf("%d checked; %d suggested", report.Checked, len(report.Suggestions)))
		for i, peer := range report.Suggestions {
			fmt.Fprintf(w, "\n  [REACHABLE] Candidate %d\n", i+1)
			fmt.Fprintln(w, "    "+peer.URI)
			cli.Field(w, "TCP connect", fmt.Sprintf("%.2f ms (not ping RTT)", peer.ConnectMS))
			cli.Field(w, "Site 7d uptime", fmt.Sprintf("%d%%", peer.Uptime))
		}
		if len(report.Suggestions) == 0 {
			cli.Text(w, "\n  [WARN] ", "No candidate passed the TCP port check. Try another nearby country; your network may block peering ports.")
		} else {
			fmt.Fprintln(w, "\n  Optional runtime connection (replace URI with a candidate):")
			fmt.Fprintln(w, "    uqdactl addPeer uri=\"URI\"")
			cli.Text(w, "  ", "On Unix, use sudo only if your admin socket requires it. Check 'uqda peers' and an overlay network test after adding a peer. For persistence, edit Peers in your existing configuration; never replace your identity.")
		}
	}
	fmt.Fprintln(w)
	cli.Text(w, "  ", report.Notice)
	fmt.Fprintln(w)
}
