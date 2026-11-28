package outline

import "testing"

// A click names the word it falls on, a run between blanks; nothing on a
// blank; and inside the selection the selection, every line of it.
func TestClicked(t *testing.T) {
	text := &Node{Kind: "text", Kids: []*Node{
		{Text: "see sh/o:24 now", Runs: []*Node{
			{Text: "see "}, {Text: "sh/o:24", Sel: true},
			{Text: " now"}}},
		{Text: "git log -p"},
	}}
	for _, c := range []struct {
		i, at int
		want  string
	}{
		{0, 2, "see"},
		{0, 6, "sh/o:24"},
		{0, 4, ""},
		{1, 6, "log"},
		{1, 99, ""},
		{2, 1, ""},
	} {
		if got := Clicked(text, c.i, c.at); got != c.want {
			t.Errorf("line %d at %d: %q, want %q", c.i, c.at, got,
				c.want)
		}
	}
}
