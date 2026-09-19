package main

import (
	"crypto/ed25519"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/Uqda/Core/src/config"
	"github.com/Uqda/Core/src/core"
)

// validateConfig checks a parsed configuration for problems that would
// otherwise only surface at daemon startup - in the case of a malformed
// MulticastInterfaces regex, as a panic (main() below compiles it with
// regexp.MustCompile) rather than a clean error. It returns a
// human-readable description of each problem found, or nil if none were.
func validateConfig(cfg *config.NodeConfig) []string {
	var problems []string

	if len(cfg.PrivateKey) != ed25519.PrivateKeySize {
		problems = append(problems, fmt.Sprintf("private key is %d bytes, expected %d", len(cfg.PrivateKey), ed25519.PrivateKeySize))
	}

	for _, raw := range cfg.Listen {
		if err := validateListenURI(raw); err != nil {
			problems = append(problems, fmt.Sprintf("Listen %q: %s", raw, err))
		}
	}

	for _, raw := range cfg.Peers {
		if err := validatePeerURI(raw); err != nil {
			problems = append(problems, fmt.Sprintf("Peers %q: %s", raw, err))
		}
	}
	for iface, peers := range cfg.InterfacePeers {
		for _, raw := range peers {
			if err := validatePeerURI(raw); err != nil {
				problems = append(problems, fmt.Sprintf("InterfacePeers[%q] %q: %s", iface, raw, err))
			}
		}
	}

	if cfg.AdminListen != "" && strings.ToLower(cfg.AdminListen) != "none" {
		if u, err := url.Parse(cfg.AdminListen); err != nil {
			problems = append(problems, fmt.Sprintf("AdminListen %q: %s", cfg.AdminListen, err))
		} else {
			switch strings.ToLower(u.Scheme) {
			case "unix", "tcp", "":
				// Recognised, or no scheme (admin.go falls back to treating
				// the whole string as a "host:port" TCP address).
			default:
				problems = append(problems, fmt.Sprintf(
					"AdminListen %q: scheme %q is not \"unix\" or \"tcp\" - the admin socket would fail to start with this value", cfg.AdminListen, u.Scheme))
			}
		}
	}

	for _, intf := range cfg.MulticastInterfaces {
		if _, err := regexp.Compile(intf.Regex); err != nil {
			problems = append(problems, fmt.Sprintf(
				"MulticastInterfaces regex %q: %s (this would panic at daemon startup via regexp.MustCompile, not just fail to match)", intf.Regex, err))
		}
	}

	return problems
}

func validateListenURI(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return err
	}
	if !core.IsValidListenScheme(u.Scheme) {
		return fmt.Errorf("unrecognised listen scheme %q", u.Scheme)
	}
	return nil
}

func validatePeerURI(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return err
	}
	if !core.IsValidPeerScheme(u.Scheme) {
		return fmt.Errorf("unrecognised peer scheme %q", u.Scheme)
	}
	return nil
}
