package main

import (
	"os"
	"strconv"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
	"golang.org/x/term"
)

// minTableWidth is the narrowest width tables are fitted to; below it,
// truncation would remove too much to be useful.
const minTableWidth = 40

// terminalWidth returns the width to fit tables to, or 0 to leave them
// unconstrained. COLUMNS overrides the detected width; output that is not a
// terminal is never constrained, so scripts keep receiving full values.
func terminalWidth(f *os.File) int {
	if v, err := strconv.Atoi(os.Getenv("COLUMNS")); err == nil && v > 0 {
		return clampWidth(v)
	}
	fd := int(f.Fd())
	if !term.IsTerminal(fd) {
		return 0
	}
	w, _, err := term.GetSize(fd)
	if err != nil || w <= 0 {
		return 0
	}
	return clampWidth(w)
}

func clampWidth(w int) int {
	if w < minTableWidth {
		return minTableWidth
	}
	return w
}

// tableFit constrains a table to width, truncating the widest cells.
func tableFit(width int) []tablewriter.Option {
	return []tablewriter.Option{
		tablewriter.WithMaxWidth(width),
		tablewriter.WithRowAutoWrap(tw.WrapTruncate),
	}
}
