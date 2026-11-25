package tty

import (
	"reflect"
	"testing"
)

// The bytes a terminal sends decode to the keys a person pressed: Esc then
// O is two keys unless a terminal's own letter follows; a mouse report, in
// either form, is one mouse key and leaks nothing; Ctrl and a letter is a
// chord by name, the keys named already keeping their names; an arrow
// with shift held is the arrow's name after S-, and with ctrl the arrow.
func TestParseKeys(t *testing.T) {
	esc, o := Key{Name: KEsc}, func(r rune) Key { return Key{Rune: r} }
	for _, c := range []struct {
		in   string
		want []Key
		rest string
	}{
		{"\x1bOA", []Key{{Name: KUp}}, ""},
		{"\x1bOxy", []Key{esc, o('O'), o('x'), o('y')}, ""},
		{"\x1bO", nil, "\x1bO"},
		{"\x1b[<0;10;5M", []Key{{Name: KMouse, X: 10, Y: 5}}, ""},
		{"\x1b[<2;3;4m",
			[]Key{{Name: KMouseUp, X: 3, Y: 4, Button: 2}}, ""},
		{"\x1b[M *%", []Key{{Name: KMouse, X: 10, Y: 5}}, ""},
		{"\x1b[M#*%", []Key{{Name: KMouseUp, X: 10, Y: 5, Button: 3}},
			""},
		{"\x1b[M *", nil, "\x1b[M *"},
		{"\x12\x03\r\t", []Key{{Name: "C-r"}, {Name: KInt},
			{Name: KEnter}, {Name: KTab}}, ""},
		{"\x1b[A", []Key{{Name: KUp}}, ""},
		{"\x1b[1;2A\x1b[1;2B\x1b[1;2D", []Key{{Name: KShiftUp},
			{Name: KShiftDown}, {Name: KShiftLeft}}, ""},
		{"\x1b[1;5A", []Key{{Name: KUp}}, ""},
	} {
		got, rest := parseKeys([]byte(c.in))
		if !reflect.DeepEqual(got, c.want) || string(rest) != c.rest {
			t.Errorf("%q: %+v rest %q, want %+v rest %q", c.in,
				got, rest, c.want, c.rest)
		}
	}
}

// Esc then O held with nothing after it is handed back as the two keys
// once the wait gives up, not dropped.
func TestFlushRest(t *testing.T) {
	term := &Term{rest: []byte("\x1bO")}
	term.flushRest()
	want := []Key{{Name: KEsc}, {Rune: 'O'}}
	if !reflect.DeepEqual(term.queue, want) || term.rest != nil {
		t.Errorf("queue %+v rest %q, want %+v", term.queue, term.rest,
			want)
	}
}
