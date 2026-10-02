package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Uqda/Core/internal/cli"
	"github.com/Uqda/Core/src/version"
)

// dispatchCLI leaves daemon flags alone; ordinary invocations use the controller
// shipped alongside this executable rather than an arbitrary binary in PATH.
func dispatchCLI(args []string, advancedHelp func()) (int, bool) {
	if len(args) > 0 && strings.HasPrefix(args[0], "-") {
		return 0, false
	}
	if len(args) == 0 {
		args = []string{"status"}
	}
	switch strings.ToLower(args[0]) {
	case "help":
		if len(args) == 2 && strings.EqualFold(args[1], "advanced") {
			advancedHelp()
		} else if len(args) == 1 {
			cli.Help(os.Stdout, "uqda", version.DisplayName())
		} else if len(args) == 2 && cli.HelpTopic(os.Stdout, "uqda", version.DisplayName(), args[1]) {
			return 0, true
		} else {
			fmt.Fprintln(os.Stderr, "Uqda: use 'uqda help', 'uqda help COMMAND' or 'uqda help advanced'.")
			return 2, true
		}
		return 0, true
	case "version":
		if len(args) != 1 {
			fmt.Fprintln(os.Stderr, "Uqda: version takes no arguments.")
			return 2, true
		}
		fmt.Println(version.DisplayName())
		return 0, true
	case "status", "doctor", "peers", "info", "test", "commands", "find":
	default:
		fmt.Fprintf(os.Stderr, "Uqda: unknown command %q. Run 'uqda help'.\n", args[0])
		return 2, true
	}
	executable, err := os.Executable()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Uqda: cannot locate the installed executable:", err)
		return 1, true
	}
	name := "uqdactl"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	controller := filepath.Join(filepath.Dir(executable), name)
	command := exec.Command(controller, args...)
	command.Stdin, command.Stdout, command.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := command.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return exitErr.ExitCode(), true
		}
		fmt.Fprintln(os.Stderr, "Uqda: cannot start the installed uqdactl; reinstall the complete Uqda package.")
		return 1, true
	}
	return 0, true
}
