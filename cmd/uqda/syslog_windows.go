//go:build windows

package main

import "github.com/gologme/log"

// go-syslog is unavailable on Windows, so retain the existing fallback to the
// standard logger when syslog output is requested.
func newSystemLogger() *log.Logger {
	return nil
}
