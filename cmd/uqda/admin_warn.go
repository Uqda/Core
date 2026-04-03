package main

import (
	"net"
	"net/url"
	"strings"

	"github.com/gologme/log"
)

// warnIfAdminListenInsecureWithoutAuth logs when the admin API is reachable beyond
// loopback over TCP without AdminAuth (see SECURITY.md).
func warnIfAdminListenInsecureWithoutAuth(adminListen, adminAuth string, logger *log.Logger) {
	if strings.TrimSpace(adminAuth) != "" {
		return
	}
	adminListen = strings.TrimSpace(adminListen)
	if adminListen == "" || strings.EqualFold(adminListen, "none") {
		return
	}
	u, err := url.Parse(adminListen)
	if err != nil {
		return
	}
	if strings.ToLower(u.Scheme) != "tcp" {
		return
	}
	host := u.Hostname()
	if host == "" {
		return
	}
	if ip := net.ParseIP(host); ip != nil {
		if ip.IsLoopback() {
			return
		}
		if ip.IsUnspecified() {
			logger.Warnln("SECURITY: AdminListen binds all interfaces but AdminAuth is empty; set AdminAuth or use a Unix socket / loopback-only listener.")
			return
		}
		logger.Warnln("SECURITY: AdminListen is on a non-loopback address but AdminAuth is empty; set AdminAuth or restrict network access.")
		return
	}
	if strings.EqualFold(host, "localhost") {
		return
	}
	logger.Warnln("SECURITY: AdminListen uses a non-loopback hostname but AdminAuth is empty; set AdminAuth or restrict network access.")
}

// warnIfUIListenNotLoopback logs when the Web UI HTTP server binds beyond loopback (dashboard proxies Admin API).
func warnIfUIListenNotLoopback(uiListen string, logger *log.Logger) {
	uiListen = strings.TrimSpace(uiListen)
	if uiListen == "" {
		return
	}
	host, _, err := net.SplitHostPort(uiListen)
	if err != nil {
		u, perr := url.Parse("http://" + uiListen)
		if perr != nil {
			return
		}
		host = u.Hostname()
	}
	if host == "" {
		return
	}
	if ip := net.ParseIP(host); ip != nil {
		if ip.IsLoopback() || ip.IsUnspecified() {
			return
		}
		logger.Warnln("SECURITY: UIListen is not on loopback; bind to localhost or place behind a reverse proxy with authentication.")
		return
	}
	if strings.EqualFold(host, "localhost") {
		return
	}
	logger.Warnln("SECURITY: UIListen hostname is not localhost; ensure the dashboard is not exposed to untrusted networks.")
}
