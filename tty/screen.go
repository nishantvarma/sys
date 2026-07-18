package tty

// Escape/colour primitives and small paint helpers. No I/O, no state — a
// command builds a frame string from these and hands it to Term.Write.

import (
	"fmt"
	"strings"
)

const (
	Home   = "\x1b[H"
	ClrEol = "\x1b[K"
	Civis  = "\x1b[?25l"
	Cnorm  = "\x1b[?25h"
	// DECSCUSR cursor shapes. CurReset restores the terminal's default.
	CurBlock = "\x1b[2 q" // steady block
	CurBeam  = "\x1b[6 q" // steady bar
	CurReset = "\x1b[0 q"
	Clear    = "\x1b[2J"
	AltOn    = "\x1b[?1049h"
	AltOff   = "\x1b[?1049l"
	MoveL    = "\x1b[D"
	Sgr0     = "\x1b[0m"
)

// Sgr wraps s in an SGR code and a reset.
func Sgr(code, s string) string { return code + s + Sgr0 }

func Bold(s string) string   { return Sgr("\x1b[1m", s) }
func Dim(s string) string    { return Sgr("\x1b[2m", s) }
func Blue(s string) string   { return Sgr("\x1b[34m", s) }
func Cyan(s string) string   { return Sgr("\x1b[36m", s) }
func Green(s string) string  { return Sgr("\x1b[32m", s) }
func Red(s string) string    { return Sgr("\x1b[31m", s) }
func Yellow(s string) string { return Sgr("\x1b[33m", s) }
func Plain(s string) string  { return s + Sgr0 }

// Line writes s, clears to end of line, and starts a new row.
func Line(b *strings.Builder, s string) {
	b.WriteString(s)
	b.WriteString(ClrEol)
	b.WriteString("\r\n")
}

// Status paints msg on row h, toggling the cursor.
func Status(h int, msg string, cursor bool) string {
	cur := Civis
	if cursor {
		cur = Cnorm
	}
	return fmt.Sprintf("\x1b[%d;1H%s%s%s", h, ClrEol, cur, msg)
}

// ScrollOffset keeps idx centred in an h-tall window over n rows.
func ScrollOffset(idx, n, h int) int {
	if n <= h {
		return 0
	}
	return Clamp(idx-h/2, 0, n-h)
}

func Clamp(x, lo, hi int) int { return max(lo, min(x, hi)) }
