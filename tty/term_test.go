package tty

import (
	"bufio"
	"strings"
	"testing"
)

// A frame writes the rows that differ from the last one, the cursor
// hidden first and put back after; none when nothing changed; all of
// them again once the screen is forgotten, as a prompt or help makes it.
func TestFrame(t *testing.T) {
	var out strings.Builder
	tm := &Term{out: bufio.NewWriter(&out)}
	frame := func(rows ...string) string {
		out.Reset()
		tm.Frame(rows, "@")
		return out.String()
	}
	row := func(n int, s string) string {
		return Goto(n, 1) + ClrEol + s
	}
	for _, c := range []struct {
		rows   []string
		forget bool
		want   string
	}{
		{[]string{"a", "b", "c"}, false, row(1, "a") + row(2, "b") +
			row(3, "c")},
		{[]string{"a", "b", "c"}, false, ""},
		{[]string{"a", "x", "c"}, false, row(2, "x")},
		{[]string{"a", "x", "c"}, true, row(1, "a") + row(2, "x") +
			row(3, "c")},
	} {
		if c.forget {
			tm.Forget()
		}
		if got := frame(c.rows...); got != Civis+c.want+"@" {
			t.Errorf("%q: wrote %q, want %q", c.rows, got,
				Civis+c.want+"@")
		}
	}
}
