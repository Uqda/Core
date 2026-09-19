package main

import (
	"strings"
	"testing"

	"github.com/Uqda/Core/src/config"
)

func TestValidateConfigAcceptsDefaults(t *testing.T) {
	cfg := config.GenerateConfig()
	if problems := validateConfig(cfg); len(problems) != 0 {
		t.Fatalf("expected a freshly generated config to be valid, got: %v", problems)
	}
}

func TestValidateConfigRejectsBadListenScheme(t *testing.T) {
	cfg := config.GenerateConfig()
	cfg.Listen = []string{"ftp://localhost:1234"}
	problems := validateConfig(cfg)
	if len(problems) != 1 || !strings.Contains(problems[0], "ftp") {
		t.Fatalf("expected exactly one problem naming the bad scheme, got: %v", problems)
	}
}

func TestValidateConfigRejectsBadPeerScheme(t *testing.T) {
	cfg := config.GenerateConfig()
	cfg.Peers = []string{"http://example.com:1234"}
	problems := validateConfig(cfg)
	if len(problems) != 1 || !strings.Contains(problems[0], "http") {
		t.Fatalf("expected exactly one problem naming the bad scheme, got: %v", problems)
	}
}

func TestValidateConfigRejectsBadInterfacePeerScheme(t *testing.T) {
	cfg := config.GenerateConfig()
	cfg.InterfacePeers = map[string][]string{"eth0": {"gopher://example.com:70"}}
	problems := validateConfig(cfg)
	if len(problems) != 1 || !strings.Contains(problems[0], "eth0") {
		t.Fatalf("expected exactly one problem naming the interface, got: %v", problems)
	}
}

func TestValidateConfigAcceptsValidPeerSchemes(t *testing.T) {
	cfg := config.GenerateConfig()
	cfg.Peers = []string{
		"tcp://a.example:1", "tls://a.example:2", "socks://a.example:3/b.example:4",
		"quic://a.example:5", "ws://a.example:6", "wss://a.example:7",
	}
	if problems := validateConfig(cfg); len(problems) != 0 {
		t.Fatalf("expected all-valid peer schemes to pass, got: %v", problems)
	}
}

func TestValidateConfigRejectsBadAdminListenScheme(t *testing.T) {
	cfg := config.GenerateConfig()
	cfg.AdminListen = "ftp://localhost:9001"
	problems := validateConfig(cfg)
	if len(problems) != 1 || !strings.Contains(problems[0], "AdminListen") {
		t.Fatalf("expected exactly one AdminListen problem, got: %v", problems)
	}
}

func TestValidateConfigAcceptsAdminListenNone(t *testing.T) {
	cfg := config.GenerateConfig()
	cfg.AdminListen = "none"
	if problems := validateConfig(cfg); len(problems) != 0 {
		t.Fatalf("expected AdminListen=none to be valid, got: %v", problems)
	}
}

func TestValidateConfigRejectsInvalidMulticastRegex(t *testing.T) {
	cfg := config.GenerateConfig()
	cfg.MulticastInterfaces = []config.MulticastInterfaceConfig{
		{Regex: "(unclosed"},
	}
	problems := validateConfig(cfg)
	if len(problems) != 1 || !strings.Contains(problems[0], "MulticastInterfaces") {
		t.Fatalf("expected exactly one MulticastInterfaces problem, got: %v", problems)
	}
	// This specific case matters because main() later does
	// regexp.MustCompile(intf.Regex) unconditionally - an invalid regex
	// here would otherwise be a panic at daemon startup, not a clean error.
	if !strings.Contains(problems[0], "panic") {
		t.Fatalf("expected the problem message to warn about the startup panic, got: %v", problems)
	}
}
