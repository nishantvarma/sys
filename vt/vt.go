// Package vt is a terminal as a host runs one, st's, ported: what a
// program writes, its text and its escape sequences, becomes a screen
// of cells, and what the host is given, a key, a paste, the mouse,
// becomes what the program reads. Its sequences mean what st's mean,
// each function of st.c that reads them one here, by st's name less the
// t most of them begin with, and its keys send what st's key table, in
// config.h, has them send. The window, the drawing, the selection and
// the printer are the host's, or no one's. tty is the other side, a
// terminal as a program here sees one.
//
// A Term is st's term and the parts of its window that the program
// sets, used from one goroutine at a time, as st's one loop uses them.
package vt

import "io"

// Term is a terminal: st's two screens, the one shown and the other, its
// cursor, and what the program set in it.
type Term struct {
	in         io.Writer // what the program reads (send)
	cols, rows int
	lines, alt [][]Cell
	c          cursor
	saved      [2]cursor // DECSC's, one a screen
	top, bot   int       // the scroll region, its first and last line
	mode       mode
	esc        esc
	csi        csi
	trantbl    [4]charset // G0 to G3
	charset    int        // the one shown, shifted to by SO and SI
	icharset   int        // the one ESC ( and its kin set
	tabs       []bool
	lastc      rune   // the last rune printed, for REP; 0 after a control
	cut        []byte // a rune the end of a Write cut, for the next
	shape      int    // the cursor's, DECSCUSR's
	buttons    int    // the mouse's held, a bit each from button 1
	ox, oy     int    // where the mouse was last said to be
}

// mode is what the program set: st's term.mode and the window's modes,
// win.mode, in one, for here the window is the host's and keeps none.
type mode uint32

const (
	modewrap mode = 1 << iota
	modeinsert
	modealtscreen
	modecrlf
	modeecho
	modeutf8
	modeappkeypad
	modemousebtn
	modemousemotion
	modereverse
	modekbdlock
	modehide
	modeappcursor
	modemousesgr
	mode8bit
	modefocus
	modemousex10
	modemousemany
	modebrcktpaste
)

const (
	// modemouse is any of the modes a program takes the mouse in.
	modemouse = modemousebtn | modemousemotion | modemousex10 |
		modemousemany
	// termmodes are st's term.mode, which a reset resets; the window's
	// stay as they were.
	termmodes = modewrap | modeinsert | modealtscreen | modecrlf |
		modeecho | modeutf8
)

// esc is how much of an escape sequence has been read, st's term.esc.
type esc uint8

const (
	escstart esc = 1 << iota
	esccsi
	escstr // DCS, OSC, PM or APC
	escaltcharset
	escstrend // a string's end came
	esctest
	escutf8
)

// cursor is st's TCursor: the attributes runes are written with, its
// rune unused, where the next one goes, and its state.
type cursor struct {
	attr  Cell
	x, y  int
	state cstate
}

type cstate uint8

const (
	cursorwrapnext cstate = 1 << iota // the next rune wraps first
	cursororigin                      // DECOM: moves count from top
)

// charset is the character set a G0 to G3 holds, of the two st knows.
type charset uint8

const (
	csusa      charset = iota
	csgraphic0         // DEC's special graphics: lines and corners
)

const (
	termname  = "st-256color" // TERM, whose terminfo says this terminal
	tabspaces = 8
	vtiden    = "\033[?6c" // what DA and DECID answer: a VT102
)

// New is a terminal cols by rows, at least a cell, what the program
// reads going to in: st's tnew, the screen blank and the modes reset.
// in must take what it is given without waiting on the program, which
// may be busy writing what the host has yet to read: st's ttywrite
// reads while it waits, and a host queues.
func New(cols, rows int, in io.Writer) *Term {
	t := &Term{in: in, c: cursor{attr: Cell{Fg: Fore, Bg: Back}}}
	t.Resize(max(1, cols), max(1, rows))
	t.reset()
	return t
}

// Write is what the program wrote, put on the screen in order, st's
// ttyread and twrite: a rune cut by the end of p waits for the rest,
// which the next Write brings. It never fails.
func (t *Term) Write(p []byte) (int, error) {
	n := len(p)
	if t.cut != nil {
		p = append(t.cut, p...)
	}
	k := t.write(p, false)
	t.cut = nil
	if k < len(p) {
		t.cut = append([]byte(nil), p[k:]...)
	}
	return n, nil
}

// Size is the terminal's columns and rows.
func (t *Term) Size() (cols, rows int) {
	return t.cols, t.rows
}

// Line is a copy of line y of the screen shown.
func (t *Term) Line(y int) []Cell {
	return append([]Cell(nil), t.lines[y]...)
}

// Cursor is the cell the cursor is drawn on: a wide rune's own, when it
// stands on the cell the rune covers, as st's draw moves it, but never
// off the line, where a DCH can leave a covered cell (setchar).
func (t *Term) Cursor() (x, y int) {
	x, y = t.c.x, t.c.y
	if t.lines[y][x].Attr&Wdummy != 0 && x > 0 {
		x--
	}
	return x, y
}

// Hidden is whether the program hid the cursor, DECTCEM.
func (t *Term) Hidden() bool {
	return t.mode&modehide != 0
}

// Shape is the cursor's shape as the program asked for it, DECSCUSR's 0
// to 7: to 2 a block, 3 and 4 an underline, 5 and 6 a bar, 7 st's
// snowman; 0 is also the host's own.
func (t *Term) Shape() int {
	return t.shape
}

// Reversed is whether the program asked for the screen in reverse,
// DECSCNM: the host's foreground and background trade places.
func (t *Term) Reversed() bool {
	return t.mode&modereverse != 0
}

// Mousing is whether the mouse, with mods held, is the program's: it
// took the mouse, and shift, st's forcemousemod, does not keep it. A
// press, a release, a move and the wheel then go to it (Press).
func (t *Term) Mousing(mods Mod) bool {
	return t.mode&modemouse != 0 && mods&Shift == 0
}
