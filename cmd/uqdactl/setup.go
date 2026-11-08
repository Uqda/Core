package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Uqda/Core/internal/publicpeers"
	"github.com/Uqda/Core/src/config"
)

type setupOptions struct {
	mode    string // private or public
	country string
	peers   int
	path    string
	confirm bool
}

type setupResult struct {
	Config        string   `json:"config"`
	Created       bool     `json:"created"`
	Mode          string   `json:"mode"`
	GroupPassword string   `json:"group_password,omitempty"` // shown once, only when generated
	PeersAdded    []string `json:"peers_added"`
	Changes       []string `json:"changes"`
	Backup        string   `json:"backup,omitempty"`
	Next          []string `json:"next"`
}

func parseSetupArgs(args []string) (setupOptions, error) {
	opts := setupOptions{peers: 2, path: config.GetDefaults().DefaultConfigFile}
	seen := map[string]bool{}
	for _, arg := range args {
		key, value, ok := strings.Cut(arg, "=")
		if !ok || value == "" || seen[key] {
			return opts, fmt.Errorf("use setup [mode=private|public] [country=NAME] [peers=0..5] [config=PATH]")
		}
		seen[key] = true
		switch key {
		case "mode":
			value = strings.ToLower(value)
			if value != "private" && value != "public" {
				return opts, fmt.Errorf("mode must be private or public")
			}
			opts.mode = value
		case "country":
			opts.country = strings.ToLower(value)
		case "peers":
			n, err := strconv.Atoi(value)
			if err != nil || n < 0 || n > 5 {
				return opts, fmt.Errorf("peers must be 0..5")
			}
			opts.peers = n
		case "config":
			opts.path = value
		case "confirm":
			opts.confirm = value == "yes"
		default:
			return opts, fmt.Errorf("unknown setup option %q", key)
		}
	}
	return opts, nil
}

func runSetup(args []string, inJSON bool) int {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	return executeSetup(ctx, args, inJSON, os.Stdout, os.Stderr, publicpeers.Fetch, publicpeers.Suggest)
}

func newGroupPassword() (string, error) {
	b := make([]byte, 24) // 192 bits
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func executeSetup(ctx context.Context, args []string, inJSON bool, out, errOut io.Writer,
	fetch func(context.Context) (publicpeers.Catalog, error),
	suggest func(context.Context, publicpeers.Catalog, string, int) (publicpeers.Report, error)) int {
	opts, err := parseSetupArgs(args)
	if err != nil {
		fmt.Fprintln(errOut, "Uqda:", err, "Run 'uqda help setup'.")
		return 2
	}
	cfg, existing, err := loadOrCreateConfig(opts.path)
	if err != nil {
		fmt.Fprintln(errOut, "Uqda:", err)
		return 1
	}
	result := setupResult{Config: opts.path, Created: existing == nil, PeersAdded: []string{}, Changes: []string{}}
	dirty := result.Created

	// Network mode. New nodes default to private; existing nodes keep theirs.
	mode := opts.mode
	if mode == "" {
		mode = "private"
		if existing != nil && cfg.GroupPassword == "" {
			mode = "public"
		}
	}
	switch {
	case mode == "private" && cfg.GroupPassword == "":
		pw, err := newGroupPassword()
		if err != nil {
			fmt.Fprintln(errOut, "Uqda:", err)
			return 1
		}
		cfg.GroupPassword = pw
		result.GroupPassword = pw
		dirty = true
		result.Changes = append(result.Changes, "created a new private network with a random group password")
	case mode == "public" && cfg.GroupPassword != "":
		if !opts.confirm {
			fmt.Fprintln(errOut, "Uqda: this node is in a private network. Switching to public removes its group protection; add confirm=yes to do it anyway.")
			return 2
		}
		cfg.GroupPassword = ""
		dirty = true
		result.Changes = append(result.Changes, "switched to the public network (group password removed)")
	}
	result.Mode = mode

	if result.Created && !cfg.Firewall.Enabled {
		cfg.Firewall = config.DefaultFirewall()
	}
	if !cfg.Firewall.Enabled && existing != nil {
		result.Next = append(result.Next, "turn on the inbound firewall: see uqda help firewall")
	}

	// Transport peers: pick healthy public TLS peers when the node has none.
	switch {
	case opts.peers == 0:
	case len(cfg.Peers) > 0:
		result.Changes = append(result.Changes, fmt.Sprintf("kept your %d existing peer(s)", len(cfg.Peers)))
	case opts.country == "":
		result.Next = append(result.Next, "add peers: run uqda find to list countries, then uqda setup country=NAME")
	default:
		catalog, err := fetch(ctx)
		if err != nil {
			fmt.Fprintln(errOut, "Uqda: could not load the public peer list:", err)
			return 1
		}
		report, err := suggest(ctx, catalog, opts.country, opts.peers)
		if err != nil {
			fmt.Fprintln(errOut, "Uqda:", err)
			return 1
		}
		for _, s := range report.Suggestions {
			cfg.Peers = append(cfg.Peers, s.URI)
			result.PeersAdded = append(result.PeersAdded, s.URI)
			dirty = true
		}
		if len(result.PeersAdded) == 0 {
			result.Next = append(result.Next, fmt.Sprintf("no reachable peers found in %s right now; try another country with uqda find", opts.country))
		} else {
			result.Changes = append(result.Changes, fmt.Sprintf("added %d fast public peer(s) in %s", len(result.PeersAdded), opts.country))
		}
		// Keep that many healthy peers from now on: the daemon replaces a
		// peer that goes down and moves to a much faster one when it appears.
		if !cfg.AutoPeers.Enabled {
			cfg.AutoPeers = config.AutoPeersConfig{Enabled: true, Country: opts.country, Count: opts.peers}
			result.Changes = append(result.Changes, "turned on automatic peer selection, so peers that go down are replaced")
			dirty = true
		}
	}
	if len(cfg.Peers) > 0 && !cfg.AutoPeers.Enabled && opts.country != "" && opts.peers > 0 && len(result.PeersAdded) == 0 {
		result.Next = append(result.Next, "to replace peers that go down automatically, add AutoPeers: { Enabled: true, Country: \""+opts.country+"\" } to the configuration")
	}

	if dirty {
		if result.Backup, err = writeConfig(opts.path, cfg, existing, ".before-setup"); err != nil {
			fmt.Fprintln(errOut, "Uqda:", err)
			return 1
		}
		result.Next = append([]string{restartHint(result.Created)}, result.Next...)
	}
	if mode == "private" {
		result.Next = append(result.Next, "add your other devices with: uqda invite")
	}
	if len(result.Changes) == 0 {
		result.Changes = append(result.Changes, "nothing to change")
	}

	if inJSON {
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		_ = enc.Encode(result)
		return 0
	}
	if result.Created {
		fmt.Fprintf(out, "Created a new Uqda node. Configuration: %s\n", opts.path)
	} else {
		fmt.Fprintf(out, "Updated %s\n", opts.path)
	}
	for _, c := range result.Changes {
		fmt.Fprintf(out, "  - %s\n", c)
	}
	for _, p := range result.PeersAdded {
		fmt.Fprintf(out, "      %s\n", p)
	}
	if result.GroupPassword != "" {
		fmt.Fprintln(out, "  Group password (shown once; uqda invite shares it securely, so you rarely need it):")
		fmt.Fprintf(out, "      %s\n", result.GroupPassword)
	}
	if result.Backup != "" {
		fmt.Fprintf(out, "  Previous configuration saved as %s\n", result.Backup)
	}
	for i, n := range result.Next {
		label := "Next:"
		if i > 0 {
			label = "Then:"
		}
		fmt.Fprintf(out, "%s %s\n", label, n)
	}
	return 0
}
