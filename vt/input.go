package vt

import (
	"bytes"
	"fmt"
	"strings"
	"unicode/utf8"
)

// Key is a key that types no text of its own, by its name.
type Key uint8

const (
	Up Key = iota + 1
	Down
	Left
	Right
	Home
	End
	Insert
	Delete
	PageUp
	PageDown
	Backspace
	Enter
	Tab
	Esc
	F1
	F2
	F3
	F4
	F5
	F6
	F7
	F8
	F9
	F10
	F11
	F12
)

// Mod is the modifiers held with a key or a press of the mouse.
type Mod uint8

const (
	Shift Mod = 1 << iota
	Ctrl
	Alt
	Super
)

// anymod matches whatever modifiers are held, st's XK_ANY_MOD.
const anymod Mod = 0xff

// keys is st's key table, config.h's key, for the keys here: what a key
// sends with what held, in a keypad or a cursor mode or either. A row is
// for its modifiers, all of them, or any, and, where it names a mode,
// for in it, +1, or out of it, -1; the first that fits is sent. The
// keypad's rows are left out, and those for F13 on, and Mod3's, for a
// host here sends none of those.
var keys = []struct {
	k                 Key
	mod               Mod
	s                 string
	appkey, appcursor int8
}{
	{Up, Shift, "\033[1;2A", 0, 0},
	{Up, Alt, "\033[1;3A", 0, 0},
	{Up, Shift | Alt, "\033[1;4A", 0, 0},
	{Up, Ctrl, "\033[1;5A", 0, 0},
	{Up, Shift | Ctrl, "\033[1;6A", 0, 0},
	{Up, Ctrl | Alt, "\033[1;7A", 0, 0},
	{Up, Shift | Ctrl | Alt, "\033[1;8A", 0, 0},
	{Up, anymod, "\033[A", 0, -1},
	{Up, anymod, "\033OA", 0, +1},
	{Down, Shift, "\033[1;2B", 0, 0},
	{Down, Alt, "\033[1;3B", 0, 0},
	{Down, Shift | Alt, "\033[1;4B", 0, 0},
	{Down, Ctrl, "\033[1;5B", 0, 0},
	{Down, Shift | Ctrl, "\033[1;6B", 0, 0},
	{Down, Ctrl | Alt, "\033[1;7B", 0, 0},
	{Down, Shift | Ctrl | Alt, "\033[1;8B", 0, 0},
	{Down, anymod, "\033[B", 0, -1},
	{Down, anymod, "\033OB", 0, +1},
	{Left, Shift, "\033[1;2D", 0, 0},
	{Left, Alt, "\033[1;3D", 0, 0},
	{Left, Shift | Alt, "\033[1;4D", 0, 0},
	{Left, Ctrl, "\033[1;5D", 0, 0},
	{Left, Shift | Ctrl, "\033[1;6D", 0, 0},
	{Left, Ctrl | Alt, "\033[1;7D", 0, 0},
	{Left, Shift | Ctrl | Alt, "\033[1;8D", 0, 0},
	{Left, anymod, "\033[D", 0, -1},
	{Left, anymod, "\033OD", 0, +1},
	{Right, Shift, "\033[1;2C", 0, 0},
	{Right, Alt, "\033[1;3C", 0, 0},
	{Right, Shift | Alt, "\033[1;4C", 0, 0},
	{Right, Ctrl, "\033[1;5C", 0, 0},
	{Right, Shift | Ctrl, "\033[1;6C", 0, 0},
	{Right, Ctrl | Alt, "\033[1;7C", 0, 0},
	{Right, Shift | Ctrl | Alt, "\033[1;8C", 0, 0},
	{Right, anymod, "\033[C", 0, -1},
	{Right, anymod, "\033OC", 0, +1},
	{Tab, Shift, "\033[Z", 0, 0},
	{Enter, Alt, "\033\r", 0, 0},
	{Enter, anymod, "\r", 0, 0},
	{Insert, Shift, "\033[4l", -1, 0},
	{Insert, Shift, "\033[2;2~", +1, 0},
	{Insert, Ctrl, "\033[L", -1, 0},
	{Insert, Ctrl, "\033[2;5~", +1, 0},
	{Insert, anymod, "\033[4h", -1, 0},
	{Insert, anymod, "\033[2~", +1, 0},
	{Delete, Ctrl, "\033[M", -1, 0},
	{Delete, Ctrl, "\033[3;5~", +1, 0},
	{Delete, Shift, "\033[2K", -1, 0},
	{Delete, Shift, "\033[3;2~", +1, 0},
	{Delete, anymod, "\033[P", -1, 0},
	{Delete, anymod, "\033[3~", +1, 0},
	{Backspace, 0, "\177", 0, 0},
	{Backspace, Alt, "\033\177", 0, 0},
	{Home, Shift, "\033[2J", 0, -1},
	{Home, Shift, "\033[1;2H", 0, +1},
	{Home, anymod, "\033[H", 0, -1},
	{Home, anymod, "\033[1~", 0, +1},
	{End, Ctrl, "\033[J", -1, 0},
	{End, Ctrl, "\033[1;5F", +1, 0},
	{End, Shift, "\033[K", -1, 0},
	{End, Shift, "\033[1;2F", +1, 0},
	{End, anymod, "\033[4~", 0, 0},
	{PageUp, Ctrl, "\033[5;5~", 0, 0},
	{PageUp, Shift, "\033[5;2~", 0, 0},
	{PageUp, anymod, "\033[5~", 0, 0},
	{PageDown, Ctrl, "\033[6;5~", 0, 0},
	{PageDown, Shift, "\033[6;2~", 0, 0},
	{PageDown, anymod, "\033[6~", 0, 0},
	{F1, 0, "\033OP", 0, 0},
	{F1, Shift, "\033[1;2P", 0, 0},
	{F1, Ctrl, "\033[1;5P", 0, 0},
	{F1, Super, "\033[1;6P", 0, 0},
	{F1, Alt, "\033[1;3P", 0, 0},
	{F2, 0, "\033OQ", 0, 0},
	{F2, Shift, "\033[1;2Q", 0, 0},
	{F2, Ctrl, "\033[1;5Q", 0, 0},
	{F2, Super, "\033[1;6Q", 0, 0},
	{F2, Alt, "\033[1;3Q", 0, 0},
	{F3, 0, "\033OR", 0, 0},
	{F3, Shift, "\033[1;2R", 0, 0},
	{F3, Ctrl, "\033[1;5R", 0, 0},
	{F3, Super, "\033[1;6R", 0, 0},
	{F3, Alt, "\033[1;3R", 0, 0},
	{F4, 0, "\033OS", 0, 0},
	{F4, Shift, "\033[1;2S", 0, 0},
	{F4, Ctrl, "\033[1;5S", 0, 0},
	{F4, Super, "\033[1;6S", 0, 0},
	{F4, Alt, "\033[1;3S", 0, 0},
	{F5, 0, "\033[15~", 0, 0},
	{F5, Shift, "\033[15;2~", 0, 0},
	{F5, Ctrl, "\033[15;5~", 0, 0},
	{F5, Super, "\033[15;6~", 0, 0},
	{F5, Alt, "\033[15;3~", 0, 0},
	{F6, 0, "\033[17~", 0, 0},
	{F6, Shift, "\033[17;2~", 0, 0},
	{F6, Ctrl, "\033[17;5~", 0, 0},
	{F6, Super, "\033[17;6~", 0, 0},
	{F6, Alt, "\033[17;3~", 0, 0},
	{F7, 0, "\033[18~", 0, 0},
	{F7, Shift, "\033[18;2~", 0, 0},
	{F7, Ctrl, "\033[18;5~", 0, 0},
	{F7, Super, "\033[18;6~", 0, 0},
	{F7, Alt, "\033[18;3~", 0, 0},
	{F8, 0, "\033[19~", 0, 0},
	{F8, Shift, "\033[19;2~", 0, 0},
	{F8, Ctrl, "\033[19;5~", 0, 0},
	{F8, Super, "\033[19;6~", 0, 0},
	{F8, Alt, "\033[19;3~", 0, 0},
	{F9, 0, "\033[20~", 0, 0},
	{F9, Shift, "\033[20;2~", 0, 0},
	{F9, Ctrl, "\033[20;5~", 0, 0},
	{F9, Super, "\033[20;6~", 0, 0},
	{F9, Alt, "\033[20;3~", 0, 0},
	{F10, 0, "\033[21~", 0, 0},
	{F10, Shift, "\033[21;2~", 0, 0},
	{F10, Ctrl, "\033[21;5~", 0, 0},
	{F10, Super, "\033[21;6~", 0, 0},
	{F10, Alt, "\033[21;3~", 0, 0},
	{F11, 0, "\033[23~", 0, 0},
	{F11, Shift, "\033[23;2~", 0, 0},
	{F11, Ctrl, "\033[23;5~", 0, 0},
	{F11, Super, "\033[23;6~", 0, 0},
	{F11, Alt, "\033[23;3~", 0, 0},
	{F12, 0, "\033[24~", 0, 0},
	{F12, Shift, "\033[24;2~", 0, 0},
	{F12, Ctrl, "\033[24;5~", 0, 0},
	{F12, Super, "\033[24;6~", 0, 0},
	{F12, Alt, "\033[24;3~", 0, 0},
}

// ascii is what the keys with no row for the modifiers held type, as X
// gives st their bytes: the four that are ASCII's.
var ascii = map[Key]rune{Backspace: '\b', Tab: '\t', Enter: '\r', Esc: 033}

// Key sends the program a key that types no text, with mods held, as
// st's kpress does: its row in the table, else, for one of ASCII's, its
// rune, typed (Type).
func (t *Term) Key(k Key, mods Mod) {
	if t.mode&modekbdlock != 0 {
		return
	}
	key := t.mode&modeappkeypad != 0
	cursor := t.mode&modeappcursor != 0
	for _, row := range keys {
		switch {
		case row.k != k, row.mod != anymod && row.mod != mods:
		case key && row.appkey < 0, !key && row.appkey > 0:
		case cursor && row.appcursor < 0, !cursor && row.appcursor > 0:
		default:
			t.send([]byte(row.s), true)
			return
		}
	}
	if r, ok := ascii[k]; ok {
		t.Type(r, mods)
	}
}

// Type sends the program a rune typed with mods held, as X and st's
// kpress give it: ctrl makes it a control code, as X makes it, and alt
// puts ESC before one byte, or, in 8-bit mode, sets its high bit.
func (t *Term) Type(r rune, mods Mod) {
	if t.mode&modekbdlock != 0 {
		return
	}
	if mods&Ctrl != 0 {
		r = ctrl(r)
	}
	b := utf8.AppendRune(nil, r)
	if len(b) == 1 && mods&Alt != 0 {
		switch {
		case t.mode&mode8bit == 0:
			b = []byte{033, b[0]}
		case b[0] < 0177:
			b = utf8.AppendRune(nil, rune(b[0])|0x80)
		}
	}
	t.send(b, true)
}

// ctrl is r typed with ctrl held, as X's XLookupString makes it: @ to ~
// and the space their control codes, 2 NUL, 3 to 7 ESC to US, 8 DEL and
// / US.
func ctrl(r rune) rune {
	switch {
	case r >= '@' && r < 0177 || r == ' ':
		return r & 0x1f
	case r == '2':
		return 0
	case r >= '3' && r <= '7':
		return r - ('3' - 033)
	case r == '8':
		return 0177
	case r == '/':
		return '_' & 0x1f
	}
	return r
}

// Paste sends the program text the host pastes, as st's selnotify does:
// each newline as the return a typist sends, and, when the program asked
// for a paste fenced, between ESC [ 200 ~ and ESC [ 201 ~.
func (t *Term) Paste(text string) {
	fenced := t.mode&modebrcktpaste != 0
	if fenced {
		t.send([]byte("\033[200~"), false)
	}
	t.send([]byte(strings.ReplaceAll(text, "\n", "\r")), true)
	if fenced {
		t.send([]byte("\033[201~"), false)
	}
}

// Focus says the host gave the terminal the eye, or took it, to the
// program that asked to know, as st's focus does.
func (t *Term) Focus(in bool) {
	switch {
	case t.mode&modefocus == 0:
	case in:
		t.send([]byte("\033[I"), false)
	default:
		t.send([]byte("\033[O"), false)
	}
}

// Press is a press of mouse button b at cell x, y with mods held: 1 to
// 3 left, middle and right, 4 and 5 the wheel up and down, 6 and 7 left
// and right, 8 to 11 more. It goes to a program that took the mouse
// (Mousing), as st's mousereport sends it.
func (t *Term) Press(b, x, y int, mods Mod) {
	if b >= 1 && b <= 11 {
		t.buttons |= 1 << (b - 1)
	}
	t.report(press, b, x, y, mods)
}

// Release is a release of mouse button b at cell x, y, as Press has it.
func (t *Term) Release(b, x, y int, mods Mod) {
	if b >= 1 && b <= 11 {
		t.buttons &^= 1 << (b - 1)
	}
	t.report(release, b, x, y, mods)
}

// Move is a move of the mouse to cell x, y, as Press has it.
func (t *Term) Move(x, y int, mods Mod) {
	t.report(motion, 0, x, y, mods)
}

// Wheel is a turn of the wheel at cell x, y, up or down, as st takes
// one: to a program that took the mouse, unless shift is held, a press
// of button 4 or 5, which no release follows; else what st's mouse
// shortcuts send, config.h's mshortcuts, ^Y or ^E, a page with shift.
func (t *Term) Wheel(up bool, x, y int, mods Mod) {
	b := 5
	if up {
		b = 4
	}
	switch {
	case t.Mousing() && mods&Shift == 0:
		t.Press(b, x, y, mods)
		t.buttons &^= 1 << (b - 1)
	case mods&Shift != 0 && up:
		t.send([]byte("\033[5;2~"), true)
	case mods&Shift != 0:
		t.send([]byte("\033[6;2~"), true)
	case up:
		t.send([]byte("\031"), true)
	default:
		t.send([]byte("\005"), true)
	}
}

// event is what the mouse did.
type event uint8

const (
	press event = iota
	release
	motion
)

// report sends the program what the mouse did, st's mousereport: a move
// only to a new cell and as the program's mode asks, with a button held
// or with any, a release but in X10's mode or of the wheel, and the
// button, the modifiers held and the cell as SGR's numbers or X10's
// bytes, which end at column and row 223.
func (t *Term) report(e event, b, x, y int, mods Mod) {
	if t.mode&modemouse == 0 {
		return
	}
	code := 0
	if e == motion {
		if x == t.ox && y == t.oy ||
			t.mode&(modemousemotion|modemousemany) == 0 ||
			t.mode&modemousemotion != 0 && t.buttons == 0 {
			return
		}
		for b = 1; b <= 11 && t.buttons&(1<<(b-1)) == 0; b++ {
		}
		code = 32
	} else if b < 1 || b > 11 || e == release &&
		(t.mode&modemousex10 != 0 || b == 4 || b == 5) {
		return
	}
	t.ox, t.oy = x, y
	switch {
	case t.mode&modemousesgr == 0 && e == release || b == 12:
		code += 3
	case b >= 8:
		code += 128 + b - 8
	case b >= 4:
		code += 64 + b - 4
	default:
		code += b - 1
	}
	if t.mode&modemousex10 == 0 {
		if mods&Shift != 0 {
			code += 4
		}
		if mods&Alt != 0 {
			code += 8
		}
		if mods&Ctrl != 0 {
			code += 16
		}
	}
	switch end := byte('M'); {
	case t.mode&modemousesgr != 0:
		if e == release {
			end = 'm'
		}
		t.send(fmt.Appendf(nil, "\033[<%d;%d;%d%c", code, x+1, y+1,
			end), false)
	case x < 223 && y < 223:
		t.send([]byte{033, '[', 'M', byte(32 + code), byte(32 + x + 1),
			byte(32 + y + 1)}, false)
	}
}

// send gives the program what it reads, st's ttywrite: with echo, once
// the program turned echo on (SRM), on the screen too, its controls
// shown, and in LNM's mode each return as a return and a newline. An
// error writing is the host's to see where it reads, as the program's
// end.
func (t *Term) send(b []byte, echo bool) {
	if echo && t.mode&modeecho != 0 {
		t.write(b, true)
	}
	if t.mode&modecrlf != 0 {
		b = bytes.ReplaceAll(b, []byte("\r"), []byte("\r\n"))
	}
	t.in.Write(b)
}
