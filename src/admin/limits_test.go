package admin

import (
	"encoding/json"
	"io"
	"net"
	"strings"
	"testing"
	"time"
)

func newLimitTestSocket(t *testing.T) *AdminSocket {
	t.Helper()
	logger, _ := newTestLogger()
	a := &AdminSocket{log: logger, handlers: make(map[string]handler)}
	if err := a.AddHandler("echo", "test", nil, func(in json.RawMessage) (interface{}, error) {
		return map[string]int{"len": len(in)}, nil
	}); err != nil {
		t.Fatal(err)
	}
	return a
}

func serveOne(t *testing.T, a *AdminSocket) net.Conn {
	t.Helper()
	server, client := net.Pipe()
	go a.handleRequest(server)
	t.Cleanup(func() { _ = client.Close() })
	return client
}

func TestAdminRejectsOversizedRequest(t *testing.T) {
	a := newLimitTestSocket(t)
	client := serveOne(t, a)
	_ = client.SetDeadline(time.Now().Add(10 * time.Second))

	big := `{"request":"echo","arguments":{"x":"` + strings.Repeat("a", maxAdminRequestBytes+1) + `"}}`
	// The server stops reading at the limit, so the write may block; do it async.
	go func() { _, _ = io.WriteString(client, big) }()

	var resp AdminSocketResponse
	if err := json.NewDecoder(client).Decode(&resp); err != nil {
		t.Fatalf("expected an error response, got %v", err)
	}
	if resp.Status != "error" || !strings.Contains(resp.Error, "exceeds") {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestAdminKeepAliveLimitIsPerRequest(t *testing.T) {
	a := newLimitTestSocket(t)
	client := serveOne(t, a)
	_ = client.SetDeadline(time.Now().Add(20 * time.Second))
	decoder := json.NewDecoder(client)

	// Several requests that together exceed the per-request limit must succeed
	// on one keepalive connection, because the limit resets per request.
	payload := strings.Repeat("b", maxAdminRequestBytes/2)
	for i := 0; i < 3; i++ {
		req := `{"request":"echo","keepalive":true,"arguments":{"x":"` + payload + `"}}` + "\n"
		go func() { _, _ = io.WriteString(client, req) }()
		var resp AdminSocketResponse
		if err := decoder.Decode(&resp); err != nil {
			t.Fatalf("request %d: %v", i, err)
		}
		if resp.Status != "success" {
			t.Fatalf("request %d failed: %+v", i, resp)
		}
	}
}

func TestAdminConnectionLimit(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	a := newLimitTestSocket(t)
	a.listener = ln
	a.done = make(chan struct{})
	go a.listen()
	defer func() { _ = a.Stop() }()

	held := make([]net.Conn, 0, maxAdminConnections)
	defer func() {
		for _, c := range held {
			_ = c.Close()
		}
	}()
	for i := 0; i < maxAdminConnections; i++ {
		c, err := net.Dial("tcp", ln.Addr().String())
		if err != nil {
			t.Fatal(err)
		}
		held = append(held, c)
	}
	// Give the accept loop time to take every held connection.
	time.Sleep(300 * time.Millisecond)

	extra, err := net.Dial("tcp", ln.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer extra.Close()
	_ = extra.SetDeadline(time.Now().Add(3 * time.Second))
	_, _ = io.WriteString(extra, `{"request":"echo"}`)
	if _, err := extra.Read(make([]byte, 1)); err == nil {
		t.Fatal("connection beyond the limit was served")
	} else if ne, ok := err.(net.Error); ok && ne.Timeout() {
		t.Fatal("connection beyond the limit was held open instead of closed")
	}

	// Freeing one slot lets a new client in.
	_ = held[0].Close()
	held = held[1:]
	time.Sleep(300 * time.Millisecond)
	ok, err := net.Dial("tcp", ln.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer ok.Close()
	_ = ok.SetDeadline(time.Now().Add(3 * time.Second))
	_, _ = io.WriteString(ok, `{"request":"echo"}`)
	var resp AdminSocketResponse
	if err := json.NewDecoder(ok).Decode(&resp); err != nil || resp.Status != "success" {
		t.Fatalf("client after freeing a slot was not served: %v %+v", err, resp)
	}
}
