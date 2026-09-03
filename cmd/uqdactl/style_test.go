package main

import (
	"bytes"
	"regexp"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
)

func TestShouldUseColor(t *testing.T) {
	tests := []struct {
		name, mode, noColor, term string
		json, tty, want           bool
		wantErr                   bool
	}{
		{name: "automatic terminal", mode: "auto", tty: true, term: "xterm-256color", want: true},
		{name: "automatic pipe", mode: "auto", tty: false, term: "xterm-256color"},
		{name: "no color convention", mode: "auto", tty: true, noColor: "1", term: "xterm-256color"},
		{name: "dumb terminal", mode: "auto", tty: true, term: "dumb"},
		{name: "explicit always", mode: "always", tty: false, noColor: "1", want: true},
		{name: "explicit never", mode: "never", tty: true},
		{name: "json is always clean", mode: "always", json: true, tty: true},
		{name: "invalid", mode: "sometimes", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getenv := func(key string) string {
				switch key {
				case "NO_COLOR":
					return tt.noColor
				case "TERM":
					return tt.term
				default:
					return ""
				}
			}
			got, err := shouldUseColor(tt.mode, tt.json, tt.tty, getenv)
			if (err != nil) != tt.wantErr {
				t.Fatalf("shouldUseColor() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("shouldUseColor() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOutputStyleStatus(t *testing.T) {
	style := outputStyle{enabled: true}
	for _, status := range []string{"PASS", "WARN", "FAIL", "NEXT", "Up", "Down"} {
		got := style.status(status)
		if !strings.HasPrefix(got, "\x1b[") || !strings.HasSuffix(got, ansiReset) {
			t.Fatalf("status %q was not colorized: %q", status, got)
		}
	}
	if got := (outputStyle{}).status("PASS"); got != "PASS" {
		t.Fatalf("disabled style changed output: %q", got)
	}
}

func TestColoredTableKeepsAlignment(t *testing.T) {
	var output bytes.Buffer
	style := outputStyle{enabled: true}
	table := tablewriter.NewTable(
		&output,
		tablewriter.WithRowAlignment(tw.AlignLeft),
		tablewriter.WithHeaderAlignment(tw.AlignCenter),
		tablewriter.WithHeaderAutoFormat(tw.Off),
	)
	table.Header(style.headers("Status", "Check", "Result"))
	_ = table.Append([]string{style.status("PASS"), style.label("daemon"), "uqda is running"})
	_ = table.Append([]string{style.status("WARNING"), style.label("peers"), "no peers are connected"})
	_ = table.Render()

	escapes := regexp.MustCompile("\\x1b\\[[0-9;]*m")
	wantWidth := 0
	for _, line := range strings.Split(strings.TrimSpace(output.String()), "\n") {
		plain := escapes.ReplaceAllString(line, "")
		width := utf8.RuneCountInString(plain)
		if wantWidth == 0 {
			wantWidth = width
		}
		if width != wantWidth {
			t.Fatalf("colored table is not aligned:\n%s\nline %q has width %d, want %d", output.String(), plain, width, wantWidth)
		}
	}
}
