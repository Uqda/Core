package cli

import (
	"io"
	"strings"

	"rsc.io/qr"
)

// QR writes text as a terminal QR code using half-block characters, with
// low error correction to keep it small enough to scan from a terminal, and
// the quiet zone that scanners need. It writes nothing if encoding fails.
func QR(w io.Writer, text string) {
	code, err := qr.Encode(text, qr.L)
	if err != nil {
		return
	}
	const quiet = 2
	size := code.Size
	dark := func(x, y int) bool {
		if x < 0 || y < 0 || x >= size || y >= size {
			return false
		}
		return code.Black(x, y)
	}
	var b strings.Builder
	for y := -quiet; y < size+quiet; y += 2 {
		for x := -quiet; x < size+quiet; x++ {
			top, bottom := dark(x, y), dark(x, y+1)
			// Light modules are printed as blocks so the code reads correctly
			// on dark terminal backgrounds.
			switch {
			case !top && !bottom:
				b.WriteRune('█')
			case !top && bottom:
				b.WriteRune('▀')
			case top && !bottom:
				b.WriteRune('▄')
			default:
				b.WriteRune(' ')
			}
		}
		b.WriteRune('\n')
	}
	_, _ = io.WriteString(w, b.String())
}
