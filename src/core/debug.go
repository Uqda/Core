//go:build debug

package core

import (
	"fmt"
	"net/http"
	_ "net/http/pprof" // Register handlers on the debug-only default mux.
	"os"
	"time"
)

// Start the profiler if the required environment variable is set.
func init() {
	envVarName := "PPROFLISTEN"
	if hostPort := os.Getenv(envVarName); hostPort != "" {
		fmt.Fprintf(os.Stderr, "DEBUG: Starting pprof on %s\n", hostPort)
		go func() {
			server := &http.Server{
				Addr:              hostPort,
				Handler:           http.DefaultServeMux,
				ReadHeaderTimeout: 5 * time.Second,
			}
			fmt.Fprintf(os.Stderr, "DEBUG: %s", server.ListenAndServe())
		}()
	}
}
