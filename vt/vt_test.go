package vt

import (
	"bytes"
	"io"
	"slices"
	"strings"
	"testing"
)

// lines is the screen as text, a line each, its trailing blanks cut and
// the cells wide runes cover left out.
func lines(t *Term) []string {
	_, rows := t.Size()
	var out []string
	for y := range rows {
		var b strings.Builder
		for _, c := range t.Line(y) {
			if c.Attr&Wdummy == 0 {
				b.WriteRune(c.Rune)
			}
		}
		out = append(out, strings.TrimRight(b.String(), " "))
	}
	return out
}

// TestScreen: what a program writes, text, control codes and sequences,
// leaves the screen and the cursor as st leaves them.
func TestScreen(t *testing.T) {
	for _, c := range []struct {
		name       string
		cols, rows int
		in         string
		want       []string
		x, y       int
	}{
		{"text wraps at a line's end", 5, 3, "abcdefg",
			[]string{"abcde", "fg", ""}, 2, 1},
		{"a return and a newline", 5, 3, "ab\r\ncd",
			[]string{"ab", "cd", ""}, 2, 1},
		{"a newline alone keeps the column", 6, 2, "ab\ncd",
			[]string{"ab", "  cd"}, 4, 1},
		{"a backspace writes over", 5, 1, "abc\bX",
			[]string{"abX"}, 3, 0},
		{"a tab goes to the next stop", 20, 1, "a\tb",
			[]string{"a       b"}, 9, 0},
		{"the last line scrolls", 3, 2, "a\r\nb\r\nc",
			[]string{"b", "c"}, 1, 1},
		{"CUP moves to a row and column", 5, 3, "\033[2;3HX",
			[]string{"", "  X", ""}, 3, 1},
		{"ED 2 blanks the screen", 5, 2, "abc\r\nde\033[2J",
			[]string{"", ""}, 2, 1},
		{"ED 1 blanks above, line 0 too", 3, 3,
			"aaa\r\nbbb\r\nccc\033[2;2H\033[1J",
			[]string{"", "  b", "ccc"}, 1, 1},
		{"EL blanks to the line's end", 5, 1, "abcde\033[3G\033[K",
			[]string{"ab"}, 2, 0},
		{"EL 1 blanks to the cursor", 5, 1, "abcde\033[3G\033[1K",
			[]string{"   de"}, 2, 0},
		{"ICH pushes the rest right", 5, 1, "abcde\033[1G\033[2@",
			[]string{"  abc"}, 0, 0},
		{"DCH pulls the rest left", 5, 1, "abcde\033[1G\033[2P",
			[]string{"cde"}, 0, 0},
		{"ECH blanks in place", 5, 1, "abcde\033[2G\033[2X",
			[]string{"a  de"}, 1, 0},
		{"IL in the scroll region", 3, 4,
			"1\r\n2\r\n3\r\n4\033[2;3r\033[2H\033[L",
			[]string{"1", "", "2", "4"}, 0, 1},
		{"DL in the scroll region", 3, 4,
			"1\r\n2\r\n3\r\n4\033[2;3r\033[2H\033[M",
			[]string{"1", "3", "", "4"}, 0, 1},
		{"RI at the top scrolls down", 3, 3, "a\r\nb\033[H\033M",
			[]string{"", "a", "b"}, 0, 0},
		{"origin mode counts from the region", 3, 4,
			"\033[2;3r\033[?6h\033[HX",
			[]string{"", "X", "", ""}, 1, 1},
		{"a wide rune DCH slid to column 0 is blanked", 5, 1,
			"中\r\033[Pa", []string{"a"}, 1, 0},
		{"the other screen keeps this one", 8, 2,
			"main\033[?1049halt\033[?1049l",
			[]string{"main", ""}, 4, 0},
		{"insert mode pushes right", 5, 1, "abc\033[1G\033[4hX",
			[]string{"Xabc"}, 1, 0},
		{"no wrap writes over the last cell", 3, 1, "\033[?7labcd",
			[]string{"abd"}, 2, 0},
		{"a wide rune takes two cells", 4, 1, "中a",
			[]string{"中a"}, 3, 0},
		{"a wide rune at the end wraps", 3, 2, "ab中",
			[]string{"ab", "中"}, 2, 1},
		{"DEC's graphics, and the USA's again", 5, 1,
			"\033(0qx\033(Bq", []string{"─│q"}, 3, 0},
		{"REP repeats the last rune", 6, 1, "a\033[3b",
			[]string{"aaaa"}, 4, 0},
		{"DECALN fills with E", 3, 2, "\033#8",
			[]string{"EEE", "EEE"}, 0, 0},
		{"no tab stops: a tab to the end", 6, 1, "\033[3g\tX",
			[]string{"     X"}, 5, 0},
	} {
		tm := New(c.cols, c.rows, io.Discard)
		tm.Write([]byte(c.in))
		if got := lines(tm); !slices.Equal(got, c.want) {
			t.Errorf("%s: the screen is %q, not %q", c.name, got,
				c.want)
		}
		if x, y := tm.Cursor(); x != c.x || y != c.y {
			t.Errorf("%s: the cursor is at %d,%d, not %d,%d",
				c.name, x, y, c.x, c.y)
		}
	}
}

// TestCut: a rune or a sequence the end of a Write cuts goes on with the
// next.
func TestCut(t *testing.T) {
	tm := New(5, 1, io.Discard)
	e := []byte("é")
	for _, p := range [][]byte{e[:1], e[1:], []byte("\033["),
		[]byte("3G"), []byte("X")} {
		tm.Write(p)
	}
	if got := lines(tm); got[0] != "é X" {
		t.Errorf("written in parts, the line is %q", got[0])
	}
}

// TestReplies: what a program asks the terminal, it answers as st does.
func TestReplies(t *testing.T) {
	for _, c := range []struct{ ask, answer string }{
		{"\033[c", "\033[?6c"},
		{"\033Z", "\033[?6c"},
		{"\033[5n", "\033[0n"},
		{"\033[3;4H\033[6n", "\033[3;4R"},
	} {
		var in bytes.Buffer
		New(5, 5, &in).Write([]byte(c.ask))
		if in.String() != c.answer {
			t.Errorf("%q is answered %q, not %q", c.ask,
				in.String(), c.answer)
		}
	}
}

// TestString: a string, an OSC's, a DCS's or the old title's, is taken
// up to its end, BEL or ST, and leaves nothing on the screen.
func TestString(t *testing.T) {
	tm := New(5, 1, io.Discard)
	tm.Write([]byte("\033]0;one\aa\033]2;two\033\\b\033Pq#0\033\\c" +
		"\033kold\033\\d"))
	if got := lines(tm); got[0] != "abcd" {
		t.Errorf("strings leave %q on the screen", got[0])
	}
}

// TestAttrs: SGR sets what runes are written with, and a colour named by
// its index or its red, green and blue is that colour.
func TestAttrs(t *testing.T) {
	tm := New(5, 1, io.Discard)
	tm.Write([]byte("\033[1;31mX\033[0mY\033[38;5;196mA" +
		"\033[48;2;1;2;3mB"))
	rgb := truecolor | 1<<16 | 2<<8 | 3
	want := []Cell{
		{'X', Bold, 1, Back},
		{'Y', 0, Fore, Back},
		{'A', 0, 196, Back},
		{'B', 0, 196, rgb},
	}
	if got := tm.Line(0)[:4]; !slices.Equal(got, want) {
		t.Errorf("the cells are %v, not %v", got, want)
	}
	for _, c := range []struct {
		c       Color
		r, g, b uint8
		ok      bool
	}{
		{196, 0xff, 0, 0, true},
		{232, 0x08, 0x08, 0x08, true},
		{255, 0xee, 0xee, 0xee, true},
		{rgb, 1, 2, 3, true},
		{1, 0, 0, 0, false},
		{Fore, 0, 0, 0, false},
	} {
		r, g, b, ok := c.c.Rgb()
		if r != c.r || g != c.g || b != c.b || ok != c.ok {
			t.Errorf("colour %d is %d,%d,%d %v", c.c, r, g, b, ok)
		}
	}
}

// TestModes: what a program sets for its host to see, the cursor hidden
// and its shape, the screen reversed and the mouse taken, a host sees.
func TestModes(t *testing.T) {
	tm := New(5, 1, io.Discard)
	for _, c := range []struct {
		in   string
		ok   func() bool
		want bool
	}{
		{"\033[?25l", tm.Hidden, true},
		{"\033[?25h", tm.Hidden, false},
		{"\033[?5h", tm.Reversed, true},
		{"\033[?1000h", func() bool { return tm.Mousing(0) }, true},
		{"", func() bool { return tm.Mousing(Shift) }, false},
		{"\033[?1000l", func() bool { return tm.Mousing(0) }, false},
	} {
		tm.Write([]byte(c.in))
		if c.ok() != c.want {
			t.Errorf("after %q, %v, not %v", c.in, c.ok(), c.want)
		}
	}
	tm.Write([]byte("\033[5 q"))
	if tm.Shape() != 5 {
		t.Errorf("DECSCUSR 5 gives shape %d", tm.Shape())
	}
}

// TestResize: a screen made shorter loses the lines above the cursor's,
// a wider one is blank past the old end, and its tab stops go on there.
func TestResize(t *testing.T) {
	tm := New(3, 3, io.Discard)
	tm.Write([]byte("a\r\nb\r\nc"))
	tm.Resize(3, 2)
	if got := lines(tm); !slices.Equal(got, []string{"b", "c"}) {
		t.Errorf("two rows from three leave %q", got)
	}
	if x, y := tm.Cursor(); x != 1 || y != 1 {
		t.Errorf("and the cursor at %d,%d", x, y)
	}
	tm = New(8, 1, io.Discard)
	tm.Resize(20, 1)
	tm.Write([]byte("\tX"))
	if got := lines(tm); got[0] != "        X" {
		t.Errorf("grown, a tab stops at %q", got[0])
	}
}

// TestReset: RIS resets the terminal's modes, insert's among them, and
// keeps those st keeps in its window, the cursor keys' among them.
func TestReset(t *testing.T) {
	var in bytes.Buffer
	tm := New(5, 1, &in)
	tm.Write([]byte("\033[?1h\033[4hab\033c"))
	tm.Key(Up, 0)
	if in.String() != "\033OA" {
		t.Errorf("after RIS up sends %q", in.String())
	}
	if got := lines(tm); got[0] != "" {
		t.Errorf("and the screen holds %q", got[0])
	}
	tm.Write([]byte("ab\rc"))
	if got := lines(tm); got[0] != "cb" {
		t.Errorf("insert mode outlived RIS: %q", got[0])
	}
}

// TestSmall: a terminal is at least a cell, though asked for none.
func TestSmall(t *testing.T) {
	tm := New(0, 0, io.Discard)
	tm.Write([]byte("ab"))
	if cols, rows := tm.Size(); cols != 1 || rows != 1 {
		t.Errorf("asked for 0 by 0, it is %d by %d", cols, rows)
	}
}
