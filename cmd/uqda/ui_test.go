package main

import (
	"encoding/json"
	"testing"
)

func TestBuildAdminRequest_envelope(t *testing.T) {
	u := &uiServer{}
	b, err := u.buildAdminRequest("getPeers", []byte(`{"auth":"tok"}`))
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	if m["request"] != "getPeers" {
		t.Fatalf("request: %v", m["request"])
	}
	if m["auth"] != "tok" {
		t.Fatalf("auth: %v", m["auth"])
	}
	switch args := m["arguments"].(type) {
	case map[string]interface{}:
		if len(args) != 0 {
			t.Fatalf("arguments: %v", args)
		}
	case string:
		var inner map[string]interface{}
		_ = json.Unmarshal([]byte(args), &inner)
		if len(inner) != 0 {
			t.Fatalf("arguments: %v", inner)
		}
	default:
		t.Fatalf("arguments type %T", m["arguments"])
	}
}

func TestBuildAdminRequest_serverAuthFallback(t *testing.T) {
	u := &uiServer{serverAuth: "srv"}
	b, err := u.buildAdminRequest("getSelf", []byte("{}"))
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]interface{}
	_ = json.Unmarshal(b, &m)
	if m["auth"] != "srv" {
		t.Fatalf("want server auth, got %v", m["auth"])
	}
}
