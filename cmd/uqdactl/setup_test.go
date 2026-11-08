package main

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/hjson/hjson-go/v4"

	"github.com/Uqda/Core/internal/publicpeers"
	"github.com/Uqda/Core/src/config"
)

func fakeCatalog(context.Context) (publicpeers.Catalog, error) { return publicpeers.Catalog{}, nil }

func fakeSuggest(uris ...string) func(context.Context, publicpeers.Catalog, string, int) (publicpeers.Report, error) {
	return func(_ context.Context, _ publicpeers.Catalog, country string, limit int) (publicpeers.Report, error) {
		r := publicpeers.Report{Country: country}
		for i, u := range uris {
			if i == limit {
				break
			}
			r.Suggestions = append(r.Suggestions, publicpeers.Suggestion{Peer: publicpeers.Peer{URI: u, Country: country}})
		}
		return r, nil
	}
}

func readConfig(t *testing.T, path string) *config.NodeConfig {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	cfg := config.GenerateConfig()
	if _, err := cfg.ReadFrom(bytes.NewReader(data)); err != nil {
		t.Fatal(err)
	}
	return cfg
}

func runSetupTest(t *testing.T, args []string, suggest func(context.Context, publicpeers.Catalog, string, int) (publicpeers.Report, error)) (int, setupResult, string) {
	t.Helper()
	var out, errOut bytes.Buffer
	code := executeSetup(context.Background(), args, true, &out, &errOut, fakeCatalog, suggest)
	var res setupResult
	if code == 0 {
		if err := json.Unmarshal(out.Bytes(), &res); err != nil {
			t.Fatalf("bad JSON: %v\n%s", err, out.String())
		}
	}
	return code, res, errOut.String()
}

func TestSetupCreatesPrivateNodeWithFirewallAndPeers(t *testing.T) {
	path := filepath.Join(t.TempDir(), "uqda.conf")
	code, res, errText := runSetupTest(t, []string{"country=germany", "config=" + path},
		fakeSuggest("tls://a.example:443", "tls://b.example:443", "tls://c.example:443"))
	if code != 0 {
		t.Fatalf("setup failed: %s", errText)
	}
	cfg := readConfig(t, path)
	if !res.Created || res.Mode != "private" || len(cfg.GroupPassword) < 32 || res.GroupPassword != cfg.GroupPassword {
		t.Fatalf("expected a new private node: %+v", res)
	}
	if !cfg.AutoPeers.Enabled || cfg.AutoPeers.Country != "germany" || cfg.AutoPeers.Count != 2 {
		t.Fatalf("setup should turn on automatic peers for the chosen country: %+v", cfg.AutoPeers)
	}
	if !cfg.Firewall.Enabled || len(cfg.Peers) != 2 {
		t.Fatalf("expected firewall and 2 peers, got firewall=%v peers=%v", cfg.Firewall.Enabled, cfg.Peers)
	}
	// Unix permission bits are not meaningful on Windows, where the
	// installer's directory ACLs protect the file.
	if info, err := os.Stat(path); err == nil && runtime.GOOS != "windows" && info.Mode().Perm()&0o077 != 0 {
		t.Fatalf("configuration is readable by others: %v", info.Mode())
	}
	if !strings.Contains(strings.Join(res.Next, " "), "uqda invite") {
		t.Fatalf("next steps should mention invites: %v", res.Next)
	}
}

func TestSetupKeepsIdentityAndExistingPeers(t *testing.T) {
	path := filepath.Join(t.TempDir(), "uqda.conf")
	cfg := config.GenerateConfig()
	cfg.Peers = []string{"tls://mine:1"}
	cfg.GroupPassword = "existing-group"
	b, _ := hjson.Marshal(cfg)
	if err := os.WriteFile(path, b, 0o600); err != nil {
		t.Fatal(err)
	}
	never := func(context.Context, publicpeers.Catalog, string, int) (publicpeers.Report, error) {
		t.Fatal("peers must not be fetched when the node already has some")
		return publicpeers.Report{}, nil
	}
	code, res, errText := runSetupTest(t, []string{"country=germany", "config=" + path}, never)
	if code != 0 {
		t.Fatalf("setup failed: %s", errText)
	}
	got := readConfig(t, path)
	if hex.EncodeToString(got.PrivateKey) != hex.EncodeToString(cfg.PrivateKey) || got.GroupPassword != "existing-group" {
		t.Fatal("setup changed the identity or group of an existing node")
	}
	if res.Created || res.GroupPassword != "" || res.Backup != "" {
		t.Fatalf("nothing should have been rewritten: %+v", res)
	}
}

func TestSetupRefusesToDropPrivacyWithoutConfirmation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "uqda.conf")
	cfg := config.GenerateConfig()
	cfg.GroupPassword = "private"
	b, _ := hjson.Marshal(cfg)
	_ = os.WriteFile(path, b, 0o600)

	if code, _, _ := runSetupTest(t, []string{"mode=public", "peers=0", "config=" + path}, fakeSuggest()); code != 2 {
		t.Fatalf("switching to public without confirm=yes should fail with 2, got %d", code)
	}
	if readConfig(t, path).GroupPassword != "private" {
		t.Fatal("group password changed without confirmation")
	}
	if code, res, _ := runSetupTest(t, []string{"mode=public", "peers=0", "confirm=yes", "config=" + path}, fakeSuggest()); code != 0 || res.Mode != "public" {
		t.Fatalf("confirmed switch failed: %d %+v", code, res)
	}
	if readConfig(t, path).GroupPassword != "" {
		t.Fatal("confirmed switch did not remove the group password")
	}
	if _, err := os.Stat(path + ".before-setup"); err != nil {
		t.Fatal("no backup was made before changing the configuration")
	}
}

func TestSetupPublicNodeWithoutCountryExplainsNextStep(t *testing.T) {
	path := filepath.Join(t.TempDir(), "uqda.conf")
	code, res, errText := runSetupTest(t, []string{"mode=public", "config=" + path}, fakeSuggest())
	if code != 0 {
		t.Fatal(errText)
	}
	if res.GroupPassword != "" || readConfig(t, path).GroupPassword != "" {
		t.Fatal("public node received a group password")
	}
	if !strings.Contains(strings.Join(res.Next, " "), "uqda find") {
		t.Fatalf("expected advice to pick a country: %v", res.Next)
	}
}

func TestSetupRejectsBadArguments(t *testing.T) {
	for _, args := range [][]string{{"mode=secret"}, {"peers=9"}, {"bogus=1"}, {"country"}, {"mode=private", "mode=public"}} {
		if code, _, _ := runSetupTest(t, args, fakeSuggest()); code != 2 {
			t.Fatalf("args %v: exit %d, want 2", args, code)
		}
	}
}
