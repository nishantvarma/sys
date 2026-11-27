package vt

import (
	"bytes"
	"testing"
)

// sent is what a terminal sends the program as do runs: a Key, a Type,
// a Paste, a press of the mouse.
func sent(in *bytes.Buffer, do func()) string {
	in.Reset()
	do()
	return in.String()
}

// TestKeys: a key sends what st's table has it send, in the modes the
// program set, a rune typed with ctrl its control code and with alt
// ESC before it, and nothing while the keyboard is locked.
func TestKeys(t *testing.T) {
	var in bytes.Buffer
	tm := New(10, 2, &in)
	key := func(k Key, m Mod) func() { return func() { tm.Key(k, m) } }
	typed := func(r rune, m Mod) func() {
		return func() { tm.Type(r, m) }
	}
	for _, c := range []struct {
		set  string // what the program set first
		do   func()
		want string
	}{
		{"", key(Up, 0), "\033[A"},
		{"", key(Up, Shift), "\033[1;2A"},
		{"", key(Up, Shift|Ctrl), "\033[1;6A"},
		{"", key(Home, 0), "\033[H"},
		{"", key(F1, 0), "\033OP"},
		{"", key(F5, Ctrl), "\033[15;5~"},
		{"", key(F1, Shift|Ctrl), ""},
		{"", key(Backspace, 0), "\177"},
		{"", key(Backspace, Alt), "\033\177"},
		{"", key(Backspace, Ctrl), "\b"},
		{"", key(Tab, 0), "\t"},
		{"", key(Tab, Shift), "\033[Z"},
		{"", key(Esc, 0), "\033"},
		{"", key(Esc, Alt), "\033\033"},
		{"", key(Enter, 0), "\r"},
		{"", key(Enter, Alt), "\033\r"},
		{"", key(Insert, 0), "\033[4h"},
		{"", key(Delete, 0), "\033[P"},
		{"", typed('a', 0), "a"},
		{"", typed('a', Ctrl), "\x01"},
		{"", typed(' ', Ctrl), "\x00"},
		{"", typed('x', Alt), "\033x"},
		{"", typed('é', Alt), "é"},
		{"\033[?1h\033=", key(Up, 0), "\033OA"},
		{"", key(Home, 0), "\033[1~"},
		{"", key(Insert, 0), "\033[2~"},
		{"", key(Delete, 0), "\033[3~"},
		{"", key(End, Ctrl), "\033[1;5F"},
		{"\033[2h", key(Up, 0), ""},
		{"", typed('a', 0), ""},
		{"\033[2l\033[20h", key(Enter, 0), "\r\n"},
		{"\033[20l\033[?1034h", typed('x', Alt), "ø"},
	} {
		tm.Write([]byte(c.set))
		if got := sent(&in, c.do); got != c.want {
			t.Errorf("after %q a key sends %q, not %q", c.set, got,
				c.want)
		}
	}
}

// TestPaste: a paste goes as typed returns, fenced once the program
// asks for a fenced paste.
func TestPaste(t *testing.T) {
	var in bytes.Buffer
	tm := New(10, 2, &in)
	paste := func() { tm.Paste("a\nb") }
	if got := sent(&in, paste); got != "a\rb" {
		t.Errorf("a paste sends %q", got)
	}
	tm.Write([]byte("\033[?2004h"))
	if got := sent(&in, paste); got != "\033[200~a\rb\033[201~" {
		t.Errorf("a fenced paste sends %q", got)
	}
}

// TestFocus: the eye gained and lost goes to a program that asked.
func TestFocus(t *testing.T) {
	var in bytes.Buffer
	tm := New(10, 2, &in)
	if got := sent(&in, func() { tm.Focus(true) }); got != "" {
		t.Errorf("unasked, the eye sends %q", got)
	}
	tm.Write([]byte("\033[?1004h"))
	if got := sent(&in, func() { tm.Focus(true) }); got != "\033[I" {
		t.Errorf("the eye gained sends %q", got)
	}
	if got := sent(&in, func() { tm.Focus(false) }); got != "\033[O" {
		t.Errorf("the eye lost sends %q", got)
	}
}

// TestMouse: a press, a release and a move go to a program that took
// the mouse, as its modes ask, in X10's bytes or SGR's numbers, at a
// cell of the screen; the wheel pages with shift alone, as st's
// shortcuts match.
func TestMouse(t *testing.T) {
	var in bytes.Buffer
	tm := New(300, 300, &in)
	press := func(b, x, y int, m Mod) func() {
		return func() { tm.Press(b, x, y, m) }
	}
	release := func(b, x, y int, m Mod) func() {
		return func() { tm.Release(b, x, y, m) }
	}
	move := func(x, y int) func() { return func() { tm.Move(x, y, 0) } }
	wheel := func(up bool, m Mod) func() {
		return func() { tm.Wheel(up, 1, 1, m) }
	}
	for _, c := range []struct {
		set  string
		do   func()
		want string
	}{
		{"", press(1, 4, 2, 0), ""},
		{"", release(1, 4, 2, 0), ""},
		{"", wheel(true, 0), "\031"},
		{"", wheel(false, Shift), "\033[6;2~"},
		{"", wheel(true, Shift|Ctrl), "\031"},
		{"\033[?1000h", press(1, 4, 2, 0), "\033[M %#"},
		{"", release(1, 4, 2, 0), "\033[M#%#"},
		{"", press(4, 4, 2, 0), "\033[M`%#"},
		{"", release(4, 4, 2, 0), ""},
		{"", press(1, 4, 2, Alt|Ctrl), "\033[M8%#"},
		{"", release(1, 4, 2, 0), "\033[M#%#"},
		{"", press(1, 230, 2, 0), ""},
		{"", release(1, 230, 2, 0), ""},
		{"\033[?1006h", press(3, 229, 0, 0), "\033[<2;230;1M"},
		{"", release(3, 229, 0, 0), "\033[<2;230;1m"},
		{"", press(1, 400, 2, 0), "\033[<0;300;3M"},
		{"", release(1, 400, 2, 0), "\033[<0;300;3m"},
		{"", move(5, 5), ""},
		{"\033[?1002h", move(6, 6), ""},
		{"", press(1, 6, 6, 0), "\033[<0;7;7M"},
		{"", move(7, 6), "\033[<32;8;7M"},
		{"", move(7, 6), ""},
		{"", release(1, 7, 6, 0), "\033[<0;8;7m"},
		{"\033[?1003h", move(8, 6), "\033[<35;9;7M"},
		{"", wheel(true, 0), "\033[<64;2;2M"},
		{"", move(3, 3), "\033[<35;4;4M"},
		{"", wheel(true, Shift), "\033[5;2~"},
		{"\033[?9h", press(1, 0, 0, Ctrl), "\033[<0;1;1M"},
		{"", release(1, 0, 0, Ctrl), ""},
	} {
		tm.Write([]byte(c.set))
		if got := sent(&in, c.do); got != c.want {
			t.Errorf("after %q the mouse sends %q, not %q",
				c.set, got, c.want)
		}
	}
}
