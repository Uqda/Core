package main

import (
	"net/http"
	"strings"

	"github.com/gologme/log"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// startMetricsServer serves Prometheus text exposition at /metrics. addr should
// usually be loopback (e.g. 127.0.0.1:9090) unless access is restricted elsewhere.
func startMetricsServer(addr string, logger *log.Logger) {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return
	}
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	srv := &http.Server{Addr: addr, Handler: mux}
	go func() {
		logger.Infof("Prometheus metrics listening on http://%s/metrics", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Errorf("metrics server: %v", err)
		}
	}()
}
