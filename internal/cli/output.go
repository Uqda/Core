// Package cli provides plain, portable terminal output for Uqda commands.
package cli

import (
	"fmt"
	"io"
	"strings"
)

const rule = "------------------------------------------------------------"

func Header(w io.Writer, title, subtitle string) {
	fmt.Fprintf(w, "\n  UQDA / %s\n", title)
	if subtitle != "" {
		fmt.Fprintf(w, "  %s\n", subtitle)
	}
	fmt.Fprintf(w, "  %s\n\n", rule)
}

func Field(w io.Writer, label, value string) {
	fmt.Fprintf(w, "  %-14s %s\n", label, value)
}

// Text wraps prose without terminal escape sequences or Unicode decorations.
func Text(w io.Writer, prefix, value string) {
	line := prefix
	for _, word := range strings.Fields(value) {
		if len(line)+len(word)+1 > 78 && line != prefix {
			fmt.Fprintln(w, line)
			line = strings.Repeat(" ", len(prefix))
		}
		if line != prefix && strings.TrimSpace(line) != "" {
			line += " "
		}
		line += word
	}
	fmt.Fprintln(w, line)
}

func Help(w io.Writer, program, release string) {
	Header(w, "COMMAND CENTER", release)
	fmt.Fprintf(w, "  Start here:  %s\n\n", program)
	for _, entry := range [][2]string{
		{program, "Node health and connection status"},
		{program + " peers", "Connected peers and traffic"},
		{program + " test IP", "Test another Uqda IPv6 address"},
		{program + " info", "Node address and public identity"},
		{program + " version", "Installed release"},
	} {
		fmt.Fprintf(w, "  %-20s %s\n", entry[0], entry[1])
	}
	fmt.Fprintf(w, "\n  Automation:  %s status --json\n", program)
	fmt.Fprintf(w, "  Advanced:    %s commands\n\n", program)
}
