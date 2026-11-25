package tty

// Escape/colour primitives and small paint helpers. No I/O, no state — a
// view builds its rows from these and hands them to Term.Frame.

import (
	"strconv"
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
	// Mouse reporting, in SGR's form. On (Term.Mouse), the terminal
	// reports a press, a move with a button held, and a release, as keys
	// (KMouse, KMouseUp).
	MouseOn  = "\x1b[?1000h\x1b[?1002h\x1b[?1006h"
	MouseOff = "\x1b[?1006l\x1b[?1002l\x1b[?1000l"
	MoveL    = "\x1b[D"
	Sgr0     = "\x1b[0m"
)

// Sgr wraps s in an SGR code and a reset.
func Sgr(code, s string) string { return code + s + Sgr0 }

func Bold(s string) string   { return Sgr("\x1b[1m", s) }
func Dim(s string) string    { return Sgr("\x1b[2m", s) }
func Yellow(s string) string { return Sgr("\x1b[33m", s) }

// Cut marks where a line was cut short. Not ~, which a path already uses
// for home.
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

// Goto moves the cursor to row and col, counted from 1.
func Goto(row, col int) string {
	return "\x1b[" + strconv.Itoa(row) + ";" + strconv.Itoa(col) + "H"
}

// Status paints msg on row h, toggling the cursor.
func Status(h int, msg string, cursor bool) string {
	cur := Civis
	if cursor {
		cur = Cnorm
	}
	return Goto(h, 1) + ClrEol + cur + msg
}
