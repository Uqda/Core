package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/mattn/go-isatty"
)

const (
	ansiReset   = "\x1b[0m"
	ansiBold    = "\x1b[1m"
	ansiRed     = "\x1b[31m"
	ansiGreen   = "\x1b[32m"
	ansiYellow  = "\x1b[33m"
	ansiCyan    = "\x1b[36m"
	ansiBoldRed = "\x1b[1;31m"
)

type outputStyle struct {
	enabled bool
}

func newOutputStyle(mode string, jsonOutput bool, output *os.File) (outputStyle, error) {
	tty := isatty.IsTerminal(output.Fd()) || isatty.IsCygwinTerminal(output.Fd())
	enabled, err := shouldUseColor(mode, jsonOutput, tty, os.Getenv)
	return outputStyle{enabled: enabled}, err
}

func shouldUseColor(mode string, jsonOutput, tty bool, getenv func(string) string) (bool, error) {
	if jsonOutput {
		return false, nil
	}
	switch strings.ToLower(mode) {
	case "always":
		return true, nil
	case "never":
		return false, nil
	case "auto", "":
		return tty && getenv("NO_COLOR") == "" && getenv("TERM") != "dumb", nil
	default:
		return false, fmt.Errorf("invalid -color value %q (use auto, always, or never)", mode)
	}
}

func (s outputStyle) paint(code, value string) string {
	if !s.enabled || value == "" {
		return value
	}
	return code + value + ansiReset
}

func (s outputStyle) headers(values ...string) []string {
	for i := range values {
		values[i] = s.header(values[i])
	}
	return values
}

func (s outputStyle) header(value string) string { return s.paint(ansiBold+ansiCyan, value) }
func (s outputStyle) good(value string) string   { return s.paint(ansiGreen, value) }
func (s outputStyle) warn(value string) string   { return s.paint(ansiYellow, value) }
func (s outputStyle) bad(value string) string    { return s.paint(ansiBoldRed, value) }
func (s outputStyle) label(value string) string  { return s.paint(ansiCyan, value) }
func (s outputStyle) strong(value string) string { return s.paint(ansiBold, value) }

func (s outputStyle) status(value string) string {
	switch strings.ToLower(value) {
	case "pass", "up", "healthy", "yes", "enabled":
		return s.good(value)
	case "warn", "warning", "next":
		return s.warn(value)
	case "fail", "failure", "error", "down", "no", "disabled":
		return s.bad(value)
	default:
		return value
	}
}

func (s outputStyle) latency(value string, milliseconds float64) string {
	switch {
	case milliseconds <= 20:
		return s.good(value)
	case milliseconds <= 70:
		return s.warn(value)
	default:
		return s.bad(value)
	}
}
