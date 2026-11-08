package main

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hjson/hjson-go/v4"

	"github.com/Uqda/Core/src/config"
	"github.com/Uqda/Core/src/invite"
)

func testToken() string {
	return invite.Token{Key: strings.Repeat("ab", 32), Addresses: []string{"192.0.2.1:1"}, Secret: strings.Repeat("A", 43)}.Encode()
}

func TestJoinUpdatesExistingConfigurationAndKeepsIdentity(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "uqda.conf")
	cfg := config.GenerateConfig()
	cfg.Peers = []string{"tls://old:1"}
	original, _ := hjson.Marshal(cfg)
	if err := os.WriteFile(path, original, 0o600); err != nil {
		t.Fatal(err)
	}
	redeem := func(context.Context, invite.Token) (invite.Grant, error) {
		return invite.Grant{GroupPassword: "group", Peers: []string{"tls://new:2"}, InviterAddress: "200::1"}, nil
	}
	var out, errOut bytes.Buffer
	if code := executeJoin(context.Background(), []string{testToken(), "config=" + path}, false, &out, &errOut, redeem); code != 0 {
		t.Fatalf("join failed (%d): %s", code, errOut.String())
	}
	data, _ := os.ReadFile(path)
	got := config.GenerateConfig()
	if _, err := got.ReadFrom(bytes.NewReader(data)); err != nil {
		t.Fatal(err)
	}
	if hex.EncodeToString(got.PrivateKey) != hex.EncodeToString(cfg.PrivateKey) {
		t.Fatal("join changed the node identity")
	}
	if got.GroupPassword != "group" || len(got.Peers) != 2 {
		t.Fatalf("grant not applied: %+v", got.Peers)
	}
	if backup, err := os.ReadFile(path + ".before-join"); err != nil || !bytes.Equal(backup, original) {
		t.Fatal("previous configuration was not backed up")
	}
	if !strings.Contains(out.String(), "Next:") {
		t.Fatalf("missing next step: %s", out.String())
	}
}

func TestJoinLeavesConfigurationUntouchedOnFailure(t *testing.T) {
	path := filepath.Join(t.TempDir(), "uqda.conf")
	redeem := func(context.Context, invite.Token) (invite.Grant, error) {
		return invite.Grant{}, errors.New("the inviting node refused: invalid or expired invite")
	}
	var out, errOut bytes.Buffer
	if code := executeJoin(context.Background(), []string{testToken(), "config=" + path}, false, &out, &errOut, redeem); code == 0 {
		t.Fatal("join reported success after a refusal")
	}
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("a configuration was written although the invite was refused")
	}
}

func TestJoinRejectsBadArguments(t *testing.T) {
	never := func(context.Context, invite.Token) (invite.Grant, error) {
		t.Fatal("redeem must not be called")
		return invite.Grant{}, nil
	}
	for _, args := range [][]string{{}, {"not-an-invite"}, {testToken(), testToken()}} {
		var out, errOut bytes.Buffer
		if code := executeJoin(context.Background(), args, false, &out, &errOut, never); code != 2 {
			t.Fatalf("args %v: exit %d, want 2", args, code)
		}
	}
}
