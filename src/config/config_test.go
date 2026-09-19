package config

import (
	"bytes"
	"encoding/json"
	"testing"
)

// ReadFrom previously sliced conf[0:2] for the BOM check without
// guarding the length, so empty or single-byte configs piped via
// -useconf panicked with index out of range.
func TestConfigReadFromEmpty(t *testing.T) {
	for _, tc := range []struct {
		name string
		body []byte
	}{
		{name: "empty", body: nil},
		{name: "single byte", body: []byte{'{'}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("ReadFrom must not panic on short input, got: %v", r)
				}
			}()
			var cfg NodeConfig
			_, _ = cfg.ReadFrom(bytes.NewReader(tc.body))
		})
	}
}

// A missing or damaged persistent identity must never become a new node.
func TestReadFromRequiresIdentity(t *testing.T) {
	for _, body := range []string{`{}`, `{"PrivateKey":""}`, `{"PrivateKey":"00"}`} {
		var cfg NodeConfig
		if _, err := cfg.ReadFrom(bytes.NewBufferString(body)); err == nil {
			t.Fatalf("accepted configuration without a valid persistent identity: %s", body)
		}
	}
}

func TestReadFromPreservesIdentity(t *testing.T) {
	original := GenerateConfig()
	body, err := json.Marshal(original)
	if err != nil {
		t.Fatal(err)
	}
	var loaded NodeConfig
	if _, err := loaded.ReadFrom(bytes.NewReader(body)); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(original.PrivateKey, loaded.PrivateKey) {
		t.Fatal("loading configuration changed identity")
	}
}
