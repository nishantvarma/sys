package tty

// Escape/colour primitives and small paint helpers. No I/O, no state — a
// command builds a frame string from these and hands it to Term.Write.

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/mattn/go-runewidth"
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
	// Bracketed paste. On, the terminal wraps a paste in 200~ and 201~ so
	// it arrives as one block rather than a run of keystrokes.
	PasteOn  = "\x1b[?2004h"
	PasteOff = "\x1b[?2004l"
	// Focus reporting. On, the terminal sends CSI I when the pane gains
	// the eye and CSI O when it loses it.
	FocusOn  = "\x1b[?1004h"
	FocusOff = "\x1b[?1004l"
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

// Cut marks where Clip and ClipLeft took text away. Not ~, which a path
// already uses for home.
const Cut = "…"

// esc returns where the escape sequence at s[i] ends, or i if none starts
// there. CSI is ESC [ params, then a final byte in @..~.
func esc(s string, i int) int {
	if s[i] != 0x1b {
		return i
	}
	j := i + 1
	if j < len(s) && s[j] == '[' {
		j++
		for j < len(s) && (s[j] < 0x40 || s[j] > 0x7e) {
			j++
		}
	}
	return min(j+1, len(s))
}

// Width is the columns s takes on screen: an escape sequence takes none,
// a wide rune two.
func Width(s string) int {
	n := 0
	for i := 0; i < len(s); {
		if j := esc(s, i); j > i {
			i = j
			continue
		}
		r, size := utf8.DecodeRuneInString(s[i:])
		n += runewidth.RuneWidth(r)
		i += size
	}
	return n
}

// Clip cuts s to w columns, so a row too long for the terminal cannot wrap
// and push the frame's top off screen. A cut row ends in Cut, still in the
// row's colour; escape sequences pass whole.
func Clip(s string, w int) string {
	if Width(s) <= w {
		return s
	}
	w -= Width(Cut)
	if w < 0 {
		return ""
	}
	n := 0
	for i := 0; i < len(s); {
		if j := esc(s, i); j > i {
			i = j
			continue
		}
		r, size := utf8.DecodeRuneInString(s[i:])
		rw := runewidth.RuneWidth(r)
		if n+rw > w {
			return s[:i] + Cut + Sgr0
		}
		n += rw
		i += size
	}
	return s
}

// ClipLeft cuts s to w columns from the front, so its end stays: a path
// keeps the directory it names. Unlike Clip, s must hold no escape
// sequences; a cut from the front would drop the one that opens a colour.
func ClipLeft(s string, w int) string {
	if Width(s) <= w {
		return s
	}
	if w < Width(Cut) {
		return ""
	}
	r := []rune(s)
	n, i := Width(Cut), len(r)
	for i > 0 && n+runewidth.RuneWidth(r[i-1]) <= w {
		i--
		n += runewidth.RuneWidth(r[i])
	}
	return Cut + string(r[i:])
}

// Line clears the row, writes s, and starts a new row. The clear comes
// first: after a row that fills the width, st and xterm hold the cursor on
// the last column, and a clear there would erase that column.
func Line(b *strings.Builder, s string) {
	b.WriteString(ClrEol)
	b.WriteString(s)
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
