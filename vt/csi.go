package vt

import (
	"fmt"
	"math"
	"unicode/utf8"
)

// csi is a CSI sequence read, st's CSIEscape: what came after ESC [, a
// ? first for DEC's private ones, its numbers, and its final byte, the
// mode, with the byte after it, for the two-byte ones.
type csi struct {
	buf  []byte
	priv bool
	arg  [escargsiz]int
	narg int
	mode [2]byte
}

const (
	escbufsiz = 128 * utf8.UTFMax // st's ESC_BUF_SIZ
	escargsiz = 16
)

// csiparse reads the sequence's numbers and its mode, st's csiparse: a
// number left out is 0, and they part by ; or, once a : comes, by :.
func (t *Term) csiparse() {
	c := &t.csi
	p := c.buf
	c.narg = 0
	if len(p) > 0 && p[0] == '?' {
		c.priv = true
		p = p[1:]
	}
	sep := byte(';')
	for len(p) > 0 {
		v, n := strtol(p)
		c.arg[c.narg] = v
		c.narg++
		p = p[n:]
		if sep == ';' && len(p) > 0 && p[0] == ':' {
			sep = ':'
		}
		if len(p) == 0 || p[0] != sep || c.narg == escargsiz {
			break
		}
		p = p[1:]
	}
	if len(p) > 0 {
		c.mode[0] = p[0]
	}
	if len(p) > 1 {
		c.mode[1] = p[1]
	}
}

// strtol is C's strtol in base 10, as st's csiparse calls it: the number
// p starts with, past blanks and a sign, and how many bytes it took,
// none when it starts with no digit. One st's int cannot hold is -1, as
// st makes a long's overflow.
func strtol(p []byte) (int, int) {
	i := 0
	for i < len(p) && (p[i] == ' ' || p[i] >= '\t' && p[i] <= '\r') {
		i++
	}
	neg := false
	if i < len(p) && (p[i] == '+' || p[i] == '-') {
		neg = p[i] == '-'
		i++
	}
	start, v, over := i, 0, false
	for ; i < len(p) && p[i] >= '0' && p[i] <= '9'; i++ {
		v = v*10 + int(p[i]-'0')
		over = over || v > math.MaxInt32
	}
	switch {
	case i == start:
		return 0, 0
	case over:
		return -1, i
	case neg:
		return -v, i
	}
	return v, i
}

// def is a sequence's number v, or d when v was left out, st's DEFAULT.
func def(v, d int) int {
	if v == 0 {
		return d
	}
	return v
}

// csihandle does what the sequence read says, st's csihandle. MC, the
// printer's, is let go, for here there is none.
func (t *Term) csihandle() {
	c := &t.csi
	switch c.mode[0] {
	case '@': // ICH
		t.insertblank(def(c.arg[0], 1))
	case 'A': // CUU
		t.moveto(t.c.x, t.c.y-def(c.arg[0], 1))
	case 'B', 'e': // CUD, VPR
		t.moveto(t.c.x, t.c.y+def(c.arg[0], 1))
	case 'c': // DA
		if c.arg[0] == 0 {
			t.send([]byte(vtiden), false)
		}
	case 'b': // REP
		if t.lastc != 0 {
			for n := limit(c.arg[0], 1, 65535); n > 0; n-- {
				t.put(t.lastc)
			}
		}
	case 'C', 'a': // CUF, HPR
		t.moveto(t.c.x+def(c.arg[0], 1), t.c.y)
	case 'D': // CUB
		t.moveto(t.c.x-def(c.arg[0], 1), t.c.y)
	case 'E': // CNL
		t.moveto(0, t.c.y+def(c.arg[0], 1))
	case 'F': // CPL
		t.moveto(0, t.c.y-def(c.arg[0], 1))
	case 'g': // TBC
		switch c.arg[0] {
		case 0:
			t.tabs[t.c.x] = false
		case 3:
			clear(t.tabs)
		}
	case 'G', '`': // CHA, HPA
		t.moveto(def(c.arg[0], 1)-1, t.c.y)
	case 'H', 'f': // CUP, HVP
		t.moveato(def(c.arg[1], 1)-1, def(c.arg[0], 1)-1)
	case 'I': // CHT
		t.puttab(def(c.arg[0], 1))
	case 'J': // ED
		switch c.arg[0] {
		case 0: // below
			t.clearregion(t.c.x, t.c.y, t.cols-1, t.c.y)
			if t.c.y < t.rows-1 {
				t.clearregion(0, t.c.y+1, t.cols-1, t.rows-1)
			}
		case 1: // above: st's > 1 leaves line 0 with the cursor on 1
			if t.c.y > 0 {
				t.clearregion(0, 0, t.cols-1, t.c.y-1)
			}
			t.clearregion(0, t.c.y, t.c.x, t.c.y)
		case 2: // all
			t.clearregion(0, 0, t.cols-1, t.rows-1)
		}
	case 'K': // EL
		switch c.arg[0] {
		case 0: // right
			t.clearregion(t.c.x, t.c.y, t.cols-1, t.c.y)
		case 1: // left
			t.clearregion(0, t.c.y, t.c.x, t.c.y)
		case 2: // all
			t.clearregion(0, t.c.y, t.cols-1, t.c.y)
		}
	case 'S': // SU
		if !c.priv {
			t.scrollup(t.top, def(c.arg[0], 1))
		}
	case 'T': // SD
		t.scrolldown(t.top, def(c.arg[0], 1))
	case 'L': // IL
		t.insertblankline(def(c.arg[0], 1))
	case 'l': // RM
		t.setmode(c.priv, false, c.arg[:c.narg])
	case 'M': // DL
		t.deleteline(def(c.arg[0], 1))
	case 'X': // ECH
		t.clearregion(t.c.x, t.c.y, t.c.x+def(c.arg[0], 1)-1, t.c.y)
	case 'P': // DCH
		t.deletechar(def(c.arg[0], 1))
	case 'Z': // CBT
		t.puttab(-def(c.arg[0], 1))
	case 'd': // VPA
		t.moveato(t.c.x, def(c.arg[0], 1)-1)
	case 'h': // SM
		t.setmode(c.priv, true, c.arg[:c.narg])
	case 'm': // SGR
		t.setattr(c.arg[:c.narg])
	case 'n': // DSR
		switch c.arg[0] {
		case 5: // its status: well
			t.send([]byte("\033[0n"), false)
		case 6: // where the cursor is, CPR
			y, x := t.c.y+1, t.c.x+1
			t.send(fmt.Appendf(nil, "\033[%d;%dR", y, x), false)
		}
	case 'r': // DECSTBM
		if !c.priv {
			top, bot := def(c.arg[0], 1), def(c.arg[1], t.rows)
			t.setscroll(top-1, bot-1)
			t.moveato(0, 0)
		}
	case 's': // DECSC, as ANSI.SYS has it
		t.savecursor()
	case 'u': // DECRC, as ANSI.SYS has it
		t.loadcursor()
	case ' ':
		if c.mode[1] == 'q' && c.arg[0] >= 0 && c.arg[0] <= 7 {
			t.shape = c.arg[0] // DECSCUSR
		}
	}
}

// setmode sets each mode in args, or resets it, st's tsetmode: DEC's
// private ones with priv, else ANSI's.
func (t *Term) setmode(priv, set bool, args []int) {
	for _, a := range args {
		if !priv {
			switch a {
			case 2: // KAM
				t.mode = modebit(t.mode, set, modekbdlock)
			case 4: // IRM
				t.mode = modebit(t.mode, set, modeinsert)
			case 12: // SRM: set, the program echoes
				t.mode = modebit(t.mode, !set, modeecho)
			case 20: // LNM
				t.mode = modebit(t.mode, set, modecrlf)
			}
			continue
		}
		switch a {
		case 1: // DECCKM
			t.mode = modebit(t.mode, set, modeappcursor)
		case 5: // DECSCNM
			t.mode = modebit(t.mode, set, modereverse)
		case 6: // DECOM
			if set {
				t.c.state |= cursororigin
			} else {
				t.c.state &^= cursororigin
			}
			t.moveato(0, 0)
		case 7: // DECAWM
			t.mode = modebit(t.mode, set, modewrap)
		case 25: // DECTCEM
			t.mode = modebit(t.mode, !set, modehide)
		case 9: // X10's mouse
			t.mode = modebit(t.mode&^modemouse, set, modemousex10)
		case 1000: // a press, a release
			t.mode = modebit(t.mode&^modemouse, set, modemousebtn)
		case 1002: // and a move with a button held
			t.mode = modebit(t.mode&^modemouse, set,
				modemousemotion)
		case 1003: // and any move
			t.mode = modebit(t.mode&^modemouse, set, modemousemany)
		case 1004: // the eye gained and lost
			t.mode = modebit(t.mode, set, modefocus)
		case 1006: // the mouse in SGR's numbers
			t.mode = modebit(t.mode, set, modemousesgr)
		case 1034: // alt the high bit
			t.mode = modebit(t.mode, set, mode8bit)
		case 1049: // the other screen, and the cursor kept, xterm's
			if set {
				t.savecursor()
			} else {
				t.loadcursor()
			}
			fallthrough
		case 47, 1047: // the other screen
			alt := t.mode&modealtscreen != 0
			if alt {
				t.clearregion(0, 0, t.cols-1, t.rows-1)
			}
			if set != alt {
				t.swapscreen()
			}
			if a != 1049 {
				break
			}
			fallthrough
		case 1048: // the cursor kept
			if set {
				t.savecursor()
			} else {
				t.loadcursor()
			}
		case 2004: // a paste fenced
			t.mode = modebit(t.mode, set, modebrcktpaste)
		}
	}
}

// modebit is m with bit set, or not, st's MODBIT.
func modebit(m mode, set bool, bit mode) mode {
	if set {
		return m | bit
	}
	return m &^ bit
}

// setattr sets the attributes runes are written with by SGR's numbers,
// each in turn, st's tsetattr.
func (t *Term) setattr(args []int) {
	a := &t.c.attr
	for i := 0; i < len(args); i++ {
		switch n := args[i]; {
		case n == 0:
			a.Attr &^= Bold | Faint | Italic | Underline | Blink |
				Reverse | Invisible | Struck
			a.Fg, a.Bg = Fore, Back
		case n == 1:
			a.Attr |= Bold
		case n == 2:
			a.Attr |= Faint
		case n == 3:
			a.Attr |= Italic
		case n == 4:
			a.Attr |= Underline
		case n == 5, n == 6: // a slow blink, a fast one
			a.Attr |= Blink
		case n == 7:
			a.Attr |= Reverse
		case n == 8:
			a.Attr |= Invisible
		case n == 9:
			a.Attr |= Struck
		case n == 22:
			a.Attr &^= Bold | Faint
		case n == 23:
			a.Attr &^= Italic
		case n == 24:
			a.Attr &^= Underline
		case n == 25:
			a.Attr &^= Blink
		case n == 27:
			a.Attr &^= Reverse
		case n == 28:
			a.Attr &^= Invisible
		case n == 29:
			a.Attr &^= Struck
		case n == 38:
			if c, ok := defcolor(args, &i); ok {
				a.Fg = c
			}
		case n == 39:
			a.Fg = Fore
		case n == 48:
			if c, ok := defcolor(args, &i); ok {
				a.Bg = c
			}
		case n == 49:
			a.Bg = Back
		case n >= 30 && n <= 37:
			a.Fg = Color(n - 30)
		case n >= 40 && n <= 47:
			a.Bg = Color(n - 40)
		case n >= 90 && n <= 97:
			a.Fg = Color(n - 90 + 8)
		case n >= 100 && n <= 107:
			a.Bg = Color(n - 100 + 8)
		}
	}
}

// defcolor is the colour SGR 38 or 48 at args[*i] names, st's tdefcolor:
// after 2 its red, green and blue, after 5 its index. *i moves past what
// it read; ok is false when they name none.
func defcolor(args []int, i *int) (c Color, ok bool) {
	if *i+1 >= len(args) {
		return 0, false
	}
	switch args[*i+1] {
	case 2:
		if *i+4 >= len(args) {
			return 0, false
		}
		r, g, b := args[*i+2], args[*i+3], args[*i+4]
		*i += 4
		if r < 0 || r > 255 || g < 0 || g > 255 || b < 0 || b > 255 {
			return 0, false
		}
		return truecolor | Color(r)<<16 | Color(g)<<8 | Color(b), true
	case 5:
		if *i+2 >= len(args) {
			return 0, false
		}
		*i += 2
		if args[*i] < 0 || args[*i] > 255 {
			return 0, false
		}
		return Color(args[*i]), true
	}
	return 0, false
}
