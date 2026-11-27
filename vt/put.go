package vt

import (
	"unicode/utf8"

	"github.com/mattn/go-runewidth"
)

// write puts each rune of p, st's twrite, and is how many bytes it took:
// a rune p ends inside of waits. With shown a control code goes on the
// screen as the caret and letter that name it, as an echo shows what was
// typed.
func (t *Term) write(p []byte, shown bool) int {
	n := 0
	for n < len(p) {
		r, size := rune(p[n]), 1
		if t.mode&modeutf8 != 0 {
			if !utf8.FullRune(p[n:]) {
				break
			}
			r, size = utf8.DecodeRune(p[n:])
		}
		if shown && control(r) {
			if r&0x80 != 0 {
				r &= 0x7f
				t.put('^')
				t.put('[')
			} else if r != '\n' && r != '\r' && r != '\t' {
				r ^= 0x40
				t.put('^')
			}
		}
		t.put(r)
		n += size
	}
	return n
}

// control is whether r is a control code, C0's or C1's, st's ISCONTROL.
func control(r rune) bool {
	return r < 0x20 || r == 0x7f || c1(r)
}

// c1 is whether r is a C1 control code, st's ISCONTROLC1.
func c1(r rune) bool {
	return r >= 0x80 && r <= 0x9f
}

// widths is how many cells a rune takes, as glibc's wcwidth, st's, gives
// it in any locale: an East Asian rune of ambiguous width one, which
// runewidth's default, read from the host's locale, makes two in a ja,
// ko or zh one.
var widths = &runewidth.Condition{StrictEmojiNeutral: true}

// put takes a rune the program wrote, st's tputc: into a string under
// way, as a control code's act, as part of a sequence, or on the screen
// at the cursor, which it moves on.
func (t *Term) put(r rune) {
	ctl := control(r)
	width := 1
	if r >= 127 && t.mode&modeutf8 != 0 && !ctl {
		width = widths.RuneWidth(r)
	}
	// a string takes every rune up to its end, which only these make
	if t.esc&escstr != 0 {
		if r != '\a' && r != 030 && r != 032 && r != 033 && !c1(r) {
			return
		}
		t.esc &^= escstart | escstr
		t.esc |= escstrend
	}
	// a control code acts at once, though it comes in a sequence
	if ctl {
		if t.mode&modeutf8 != 0 && c1(r) {
			return
		}
		t.controlcode(byte(r))
		if t.esc == 0 {
			t.lastc = 0
		}
		return
	}
	if t.esc&escstart != 0 {
		switch {
		case t.esc&esccsi != 0:
			t.csi.buf = append(t.csi.buf, byte(r))
			full := len(t.csi.buf) >= escbufsiz-1
			if r >= 0x40 && r <= 0x7e || full {
				t.esc = 0
				t.csiparse()
				t.csihandle()
			}
			return
		case t.esc&escutf8 != 0:
			t.defutf8(r)
		case t.esc&escaltcharset != 0:
			t.deftran(r)
		case t.esc&esctest != 0:
			t.dectest(r)
		default:
			if !t.eschandle(r) {
				return
			}
		}
		t.esc = 0
		return
	}
	line := t.lines[t.c.y]
	if t.mode&modewrap != 0 && t.c.state&cursorwrapnext != 0 {
		t.newline(true)
		line = t.lines[t.c.y]
	}
	if t.mode&modeinsert != 0 && t.c.x+width < t.cols {
		copy(line[t.c.x+width:], line[t.c.x:t.cols-width])
		line[t.c.x].Attr &^= Wide
	}
	if t.c.x+width > t.cols {
		if t.mode&modewrap != 0 {
			t.newline(true)
		} else {
			t.moveto(t.cols-width, t.c.y)
		}
		line = t.lines[t.c.y]
	}
	t.setchar(r, t.c.attr, t.c.x, t.c.y)
	t.lastc = r
	if width == 2 {
		line[t.c.x].Attr |= Wide
		if x := t.c.x + 1; x < t.cols {
			if line[x].Attr == Wide && x+1 < t.cols {
				line[x+1].Rune = ' '
				line[x+1].Attr &^= Wdummy
			}
			line[x].Rune, line[x].Attr = 0, Wdummy
		}
	}
	if t.c.x+width < t.cols {
		t.moveto(t.c.x+width, t.c.y)
	} else {
		t.c.state |= cursorwrapnext
	}
}

// controlcode does what control code c does, st's tcontrolcode. A string
// under way ends at any but the ones that return here.
func (t *Term) controlcode(c byte) {
	switch c {
	case '\t':
		t.puttab(1)
		return
	case '\b':
		t.moveto(t.c.x-1, t.c.y)
		return
	case '\r':
		t.moveto(0, t.c.y)
		return
	case '\f', '\v', '\n':
		t.newline(t.mode&modecrlf != 0)
		return
	case '\a':
		// BEL ends a string as ST does, xterm's way; st's bell is its
		// window's, and here the window is the host's
		if t.esc&escstrend != 0 {
			t.strhandle()
		}
	case 033:
		t.csi = csi{}
		t.esc &^= esccsi | escaltcharset | esctest
		t.esc |= escstart
		return
	case 016, 017: // SO and SI: G1 shown, or G0
		t.charset = 1 - int(c-016)
		return
	case 032: // SUB
		t.setchar('?', t.c.attr, t.c.x, t.c.y)
		t.csi = csi{}
	case 030: // CAN
		t.csi = csi{}
	case 005, 000, 021, 023, 0177: // ENQ, NUL, XON, XOFF, DEL
		return
	case 0x85: // NEL
		t.newline(true)
	case 0x88: // HTS
		t.tabs[t.c.x] = true
	case 0x9a: // DECID
		t.send([]byte(vtiden), false)
	case 0x90, 0x9d, 0x9e, 0x9f: // DCS, OSC, PM, APC
		t.strsequence()
		return
	}
	t.esc &^= escstrend | escstr
}

// eschandle does what ESC and the rune after it do, st's eschandle, and
// is whether the sequence is done; false, more of it is to come.
func (t *Term) eschandle(r rune) bool {
	switch r {
	case '[':
		t.esc |= esccsi
		return false
	case '#':
		t.esc |= esctest
		return false
	case '%':
		t.esc |= escutf8
		return false
	case 'P', '_', '^', ']', 'k': // DCS, APC, PM, OSC, the old title
		t.strsequence()
		return false
	case 'n', 'o': // LS2, LS3
		t.charset = 2 + int(r-'n')
	case '(', ')', '*', '+': // the set G0 to G3 holds, to come
		t.icharset = int(r - '(')
		t.esc |= escaltcharset
		return false
	case 'D': // IND
		if t.c.y == t.bot {
			t.scrollup(t.top, 1)
		} else {
			t.moveto(t.c.x, t.c.y+1)
		}
	case 'E': // NEL
		t.newline(true)
	case 'H': // HTS
		t.tabs[t.c.x] = true
	case 'M': // RI
		if t.c.y == t.top {
			t.scrolldown(t.top, 1)
		} else {
			t.moveto(t.c.x, t.c.y-1)
		}
	case 'Z': // DECID
		t.send([]byte(vtiden), false)
	case 'c': // RIS: st reloads its colours too, and here they are the
		// host's
		t.reset()
		t.mode &^= modehide
	case '=': // DECPAM
		t.mode |= modeappkeypad
	case '>': // DECPNM
		t.mode &^= modeappkeypad
	case '7': // DECSC
		t.savecursor()
	case '8': // DECRC
		t.loadcursor()
	case '\\': // ST
		if t.esc&escstrend != 0 {
			t.strhandle()
		}
	}
	return true
}

// defutf8 is ESC % and its letter, st's tdefutf8: G turns UTF-8 on, @
// off.
func (t *Term) defutf8(r rune) {
	switch r {
	case 'G':
		t.mode |= modeutf8
	case '@':
		t.mode &^= modeutf8
	}
}

// deftran is the letter after ESC ( or its kin, st's tdeftran: the set
// the G it names holds, 0 DEC's graphics or B the USA's, the only two
// st knows.
func (t *Term) deftran(r rune) {
	switch r {
	case '0':
		t.trantbl[t.icharset] = csgraphic0
	case 'B':
		t.trantbl[t.icharset] = csusa
	}
}

// dectest is ESC # and its digit, st's tdectest: 8 fills the screen
// with E, DECALN.
func (t *Term) dectest(r rune) {
	if r != '8' {
		return
	}
	for x := range t.cols {
		for y := range t.rows {
			t.setchar('E', t.c.attr, x, y)
		}
	}
}
