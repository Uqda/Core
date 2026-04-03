package main

import (
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gologme/log"

	"github.com/Uqda/Core/src/admin"
)

const uiMaxBodyBytes = 1 << 20

type uiServer struct {
	adminListen   string
	metricsListen string
	serverAuth    string
	logger        *log.Logger
}

func startUIServer(uiListen, adminListen, metricsListen, serverAuth string, logger *log.Logger) {
	uiListen = strings.TrimSpace(uiListen)
	if uiListen == "" {
		return
	}
	u := &uiServer{
		adminListen:   adminListen,
		metricsListen: strings.TrimSpace(metricsListen),
		serverAuth:    serverAuth,
		logger:        logger,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", u.serveIndex)
	mux.HandleFunc("/health", u.serveHealth)
	mux.HandleFunc("/metrics-proxy", u.serveMetricsProxy)
	mux.HandleFunc("/api/", u.proxyAPI)

	srv := &http.Server{
		Addr:         uiListen,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	go func() {
		u.logger.Infof("Web UI listening on http://%s", uiListen)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			u.logger.Errorf("Web UI server: %v", err)
		}
	}()
}

func (u *uiServer) serveIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(uiIndexHTML)
}

func (u *uiServer) serveHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}

func (u *uiServer) serveMetricsProxy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if u.metricsListen == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "metrics not configured"})
		return
	}
	addr := strings.TrimSpace(u.metricsListen)
	addr = strings.TrimPrefix(addr, "http://")
	addr = strings.TrimPrefix(addr, "https://")
	metricsURL := "http://" + strings.TrimSuffix(addr, "/") + "/metrics"
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(metricsURL)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	defer func() { _ = resp.Body.Close() }()
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

func (u *uiServer) proxyAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	command := strings.TrimPrefix(r.URL.Path, "/api/")
	command = strings.Trim(command, "/")
	if command == "" {
		http.Error(w, "missing command", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, uiMaxBodyBytes))
	if err != nil {
		http.Error(w, "error reading body", http.StatusBadRequest)
		return
	}

	reqBytes, err := u.buildAdminRequest(command, body)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	conn, err := dialAdminSocket(u.adminListen, 5*time.Second)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status": "error",
			"error":  "cannot reach admin socket: " + err.Error(),
		})
		return
	}
	defer func() { _ = conn.Close() }()
	_ = conn.SetDeadline(time.Now().Add(30 * time.Second))

	if _, err := conn.Write(append(reqBytes, '\n')); err != nil {
		http.Error(w, "error writing to admin", http.StatusInternalServerError)
		return
	}

	dec := json.NewDecoder(io.LimitReader(conn, uiMaxBodyBytes))
	var resp json.RawMessage
	if err := dec.Decode(&resp); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "error", "error": err.Error()})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(resp)
}

// buildAdminRequest merges POST JSON with the Admin API envelope (request, arguments, auth).
func (u *uiServer) buildAdminRequest(command string, body []byte) ([]byte, error) {
	var raw map[string]interface{}
	if len(body) > 0 {
		if err := json.Unmarshal(body, &raw); err != nil {
			raw = map[string]interface{}{}
		}
	} else {
		raw = map[string]interface{}{}
	}

	auth, _ := raw["auth"].(string)
	delete(raw, "auth")
	if auth == "" && u.serverAuth != "" {
		auth = u.serverAuth
	}

	argBytes, err := json.Marshal(raw)
	if err != nil || len(argBytes) == 0 || string(argBytes) == "null" {
		argBytes = []byte("{}")
	}

	req := admin.AdminSocketRequest{
		Name:      command,
		Arguments: json.RawMessage(argBytes),
		Auth:      auth,
	}
	return json.Marshal(req)
}

func dialAdminSocket(adminListen string, timeout time.Duration) (net.Conn, error) {
	adminListen = strings.TrimSpace(adminListen)
	if adminListen == "" || strings.EqualFold(adminListen, "none") {
		return nil, io.EOF
	}
	u, err := url.Parse(adminListen)
	if err == nil && u.Scheme != "" {
		switch strings.ToLower(u.Scheme) {
		case "unix":
			path := u.Path
			if path == "" {
				path = u.Host
			}
			return net.DialTimeout("unix", path, timeout)
		case "tcp":
			return net.DialTimeout("tcp", u.Host, timeout)
		default:
			return net.DialTimeout("tcp", u.Host, timeout)
		}
	}
	return net.DialTimeout("tcp", adminListen, timeout)
}
