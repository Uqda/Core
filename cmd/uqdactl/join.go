package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/hjson/hjson-go/v4"

	"github.com/Uqda/Core/src/config"
	"github.com/Uqda/Core/src/invite"
)

type joinResult struct {
	Config         string   `json:"config"`
	Created        bool     `json:"created"`
	Changes        []string `json:"changes"`
	InviterAddress string   `json:"inviter_address"`
	Name           string   `json:"name,omitempty"`
	Backup         string   `json:"backup,omitempty"`
	Next           string   `json:"next"`
}

func runJoin(args []string, inJSON bool) int {
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	return executeJoin(ctx, args, inJSON, os.Stdout, os.Stderr, invite.Redeem)
}

func executeJoin(ctx context.Context, args []string, inJSON bool, out, errOut io.Writer,
	redeem func(context.Context, invite.Token) (invite.Grant, error)) int {
	var tokenText string
	path := config.GetDefaults().DefaultConfigFile
	for _, arg := range args {
		switch {
		case strings.HasPrefix(arg, "config="):
			path = strings.TrimPrefix(arg, "config=")
		case tokenText == "":
			tokenText = arg
		default:
			fmt.Fprintln(errOut, "Uqda: join takes one invite and an optional config=PATH. Run 'uqda help join'.")
			return 2
		}
	}
	if tokenText == "" {
		fmt.Fprintln(errOut, "Uqda: paste the invite after join, for example: uqda join uqda-invite-1:...")
		return 2
	}
	token, err := invite.Decode(tokenText)
	if err != nil {
		fmt.Fprintln(errOut, "Uqda:", err)
		return 2
	}

	cfg, existing, err := loadOrCreateConfig(path)
	if err != nil {
		fmt.Fprintln(errOut, "Uqda:", err)
		return 1
	}
	grant, err := redeem(ctx, token)
	if err != nil {
		fmt.Fprintln(errOut, "Uqda:", err)
		return 1
	}
	result := joinResult{Config: path, Created: existing == nil, InviterAddress: grant.InviterAddress, Name: grant.Name}
	result.Changes = invite.Apply(cfg, grant)
	if len(result.Changes) == 0 {
		result.Changes = []string{"already up to date"}
	}
	if result.Backup, err = writeConfig(path, cfg, existing); err != nil {
		fmt.Fprintln(errOut, "Uqda:", err)
		return 1
	}
	result.Next = restartHint(result.Created)

	if inJSON {
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		_ = enc.Encode(result)
		return 0
	}
	title := "Joined"
	if result.Name != "" {
		title += " \"" + result.Name + "\""
	}
	fmt.Fprintf(out, "%s. Configuration: %s\n", title, path)
	for _, c := range result.Changes {
		fmt.Fprintf(out, "  - %s\n", c)
	}
	if result.Backup != "" {
		fmt.Fprintf(out, "  Previous configuration saved as %s\n", result.Backup)
	}
	if grant.InviterAddress != "" {
		fmt.Fprintf(out, "  Inviting node: %s (try: uqda test %s)\n", grant.InviterAddress, grant.InviterAddress)
	}
	fmt.Fprintf(out, "Next: %s\n", result.Next)
	return 0
}

// loadOrCreateConfig returns the configuration at path, or a new one with a
// fresh identity and the default firewall when path does not exist. existing
// holds the original bytes, or nil for a new configuration.
func loadOrCreateConfig(path string) (*config.NodeConfig, []byte, error) {
	data, err := os.ReadFile(path)
	switch {
	case errors.Is(err, os.ErrNotExist):
		cfg := config.GenerateConfig()
		cfg.AdminListen = ""
		cfg.Firewall = config.DefaultFirewall()
		return cfg, nil, nil
	case err != nil:
		return nil, nil, fmt.Errorf("cannot read %s: %w (on macOS and Linux, run with sudo)", path, err)
	}
	cfg := config.GenerateConfig()
	if _, err := cfg.ReadFrom(bytes.NewReader(data)); err != nil {
		return nil, nil, fmt.Errorf("cannot parse %s: %w", path, err)
	}
	return cfg, data, nil
}

// writeConfig saves cfg atomically with owner-only permissions, keeping a
// copy of the previous file. It returns the backup path, if one was made.
func writeConfig(path string, cfg *config.NodeConfig, existing []byte) (string, error) {
	if cfg.PrivateKeyPath != "" {
		cfg.PrivateKey = nil // keep the key in its external file
	}
	encoded, err := hjson.Marshal(cfg)
	if err != nil {
		return "", fmt.Errorf("encode configuration: %w", err)
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create %s: %w", dir, err)
	}
	backup := ""
	if existing != nil {
		backup = path + ".before-join"
		if err := os.WriteFile(backup, existing, 0o600); err != nil {
			return "", fmt.Errorf("save backup %s: %w", backup, err)
		}
	}
	tmp, err := os.CreateTemp(dir, ".uqda-join-*")
	if err != nil {
		return "", fmt.Errorf("write %s: %w (on macOS and Linux, run with sudo)", path, err)
	}
	defer os.Remove(tmp.Name())
	if err := tmp.Chmod(0o600); err != nil && runtime.GOOS != "windows" {
		_ = tmp.Close()
		return "", err
	}
	if _, err := tmp.Write(append(encoded, '\n')); err != nil {
		_ = tmp.Close()
		return "", err
	}
	if err := tmp.Close(); err != nil {
		return "", err
	}
	if err := os.Rename(tmp.Name(), path); err != nil {
		return "", fmt.Errorf("replace %s: %w", path, err)
	}
	return backup, nil
}

func restartHint(created bool) string {
	switch runtime.GOOS {
	case "linux":
		if created {
			return "start Uqda: sudo systemctl enable --now uqda"
		}
		return "restart Uqda: sudo systemctl restart uqda"
	case "darwin":
		return "restart Uqda: sudo launchctl kickstart -k system/io.github.uqda.core"
	case "windows":
		return "restart Uqda from an administrator PowerShell: Restart-Service Uqda"
	default:
		return "restart the Uqda service so the new settings take effect"
	}
}
