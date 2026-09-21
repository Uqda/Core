package main

import (
	"io"
	"log"
	"strings"
	"testing"
)

func TestDialAdminEndpointRejectsMalformedAndUnsupportedEndpoints(t *testing.T) {
	logger := log.New(io.Discard, "", 0)
	for _, endpoint := range []string{"tcp://%", "http://127.0.0.1:9001", "tcp://"} {
		if conn, err := dialAdminEndpoint(endpoint, logger); err == nil {
			_ = conn.Close()
			t.Fatalf("accepted invalid endpoint %q", endpoint)
		}
	}
}

func TestDialAdminEndpointReportsUnavailableDaemon(t *testing.T) {
	logger := log.New(io.Discard, "", 0)
	_, err := dialAdminEndpoint("tcp://127.0.0.1:0", logger)
	if err == nil {
		t.Fatal("connected to an unavailable admin endpoint")
	}
	if !strings.Contains(err.Error(), "connect to admin endpoint") {
		t.Fatalf("unexpected error: %v", err)
	}
}
