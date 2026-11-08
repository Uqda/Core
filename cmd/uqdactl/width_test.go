package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
)

func TestTablesFitNarrowScreens(t *testing.T) {
	for _, width := range []int{40, 60, 80} {
		var buf bytes.Buffer
		opts := append([]tablewriter.Option{
			tablewriter.WithRenderer(renderer.NewBlueprint(tw.Rendition{Symbols: tw.NewSymbols(tw.StyleASCII)})),
		}, tableFit(width)...)
		table := tablewriter.NewTable(&buf, opts...)
		table.Header([]string{"URI", "State", "Dir", "IP Address", "Uptime"})
		_ = table.Append([]string{"tls://some-very-long-public-peer-hostname.example.org:443?key=" + strings.Repeat("a", 64), "Up", "Out", "201:7b0f:911:8803:fb15:5dec:d891:16cb", "3h2m"})
		_ = table.Render()
		for _, line := range strings.Split(strings.TrimRight(buf.String(), "\n"), "\n") {
			if n := utf8.RuneCountInString(line); n > width {
				t.Fatalf("width %d: line is %d characters:\n%s", width, n, buf.String())
			}
		}
	}
}

func TestTerminalWidthRespectsColumnsAndPipes(t *testing.T) {
	t.Setenv("COLUMNS", "72")
	if got := terminalWidth(os.Stdout); got != 72 {
		t.Fatalf("COLUMNS=72 gave %d", got)
	}
	t.Setenv("COLUMNS", "10")
	if got := terminalWidth(os.Stdout); got != minTableWidth {
		t.Fatalf("tiny COLUMNS should clamp to %d, got %d", minTableWidth, got)
	}
	t.Setenv("COLUMNS", "")
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	defer w.Close()
	if got := terminalWidth(w); got != 0 {
		t.Fatalf("a pipe must not be constrained, got %d", got)
	}
}
