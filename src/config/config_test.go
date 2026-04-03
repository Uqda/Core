package config

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestConfig_Keys(t *testing.T) {
	/*
		var nodeConfig NodeConfig
		nodeConfig.NewKeys()

		publicKey1, err := hex.DecodeString(nodeConfig.PublicKey)

		if err != nil {
			t.Fatal("can not decode generated public key")
		}

		if len(publicKey1) == 0 {
			t.Fatal("empty public key generated")
		}

		privateKey1, err := hex.DecodeString(nodeConfig.PrivateKey)

		if err != nil {
			t.Fatal("can not decode generated private key")
		}

		if len(privateKey1) == 0 {
			t.Fatal("empty private key generated")
		}

		nodeConfig.NewKeys()

		publicKey2, err := hex.DecodeString(nodeConfig.PublicKey)

		if err != nil {
			t.Fatal("can not decode generated public key")
		}

		if bytes.Equal(publicKey2, publicKey1) {
			t.Fatal("same public key generated")
		}

		privateKey2, err := hex.DecodeString(nodeConfig.PrivateKey)

		if err != nil {
			t.Fatal("can not decode generated private key")
		}

		if bytes.Equal(privateKey2, privateKey1) {
			t.Fatal("same private key generated")
		}
	*/
}

func TestPrivateKeyPath_LoadsFromPEMFile(t *testing.T) {
	base := GenerateConfig()
	if err := base.postprocessConfig(); err != nil {
		t.Fatal(err)
	}
	pemBytes, err := base.MarshalPEMPrivateKey()
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	keyPath := filepath.Join(dir, "key.pem")
	if err := os.WriteFile(keyPath, pemBytes, 0600); err != nil {
		t.Fatal(err)
	}
	cfg := GenerateConfig()
	cfg.PrivateKey = nil
	cfg.PrivateKeyPath = keyPath
	if err := cfg.postprocessConfig(); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(cfg.PrivateKey, base.PrivateKey) {
		t.Fatal("loaded private key does not match PEM file")
	}
}
