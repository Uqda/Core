package main

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"
)

// Only a denied local Unix socket connection warrants sudo guidance. Preserve
// the original error for errors.Is, without treating TCP failures as elevation
// problems or exposing invocation arguments (which can contain peer passwords).
type adminAccessError struct{ cause error }

func (e *adminAccessError) Error() string {
	return "permission denied accessing the local admin socket"
}
func (e *adminAccessError) Unwrap() error { return e.cause }

func isAdminAccessError(err error) bool {
	var denied *adminAccessError
	return errors.As(err, &denied)
}

func adminAccessHint(command string) string {
	return accessHint(command, runtime.GOOS, os.Geteuid())
}

func accessHint(command, platform string, uid int) string {
	if platform == "windows" {
		return "Check your account's access to the local admin socket; sudo is not a Windows command."
	}
	if uid == 0 {
		return "Already running as root: check socket ownership, directory access and OS security policy. Do not make the admin socket world-accessible."
	}
	example := "sudo uqda status"
	switch strings.ToLower(command) {
	case "getself", "info":
		example = "sudo uqda info"
	case "getpeers", "peers":
		example = "sudo uqda peers"
	case "list", "commands":
		example = "sudo uqda commands"
	case "doctor":
		example = "sudo uqdactl doctor"
	case "test":
		example = "sudo uqda test IP"
	case "getpaths":
		example = "sudo uqdactl getPaths"
	case "getsessions":
		example = "sudo uqdactl getSessions"
	}
	return fmt.Sprintf("Run the same command with sudo (example: %s); keep any custom endpoint and arguments. Do not make the admin socket world-accessible.", example)
}
