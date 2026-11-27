package tabs

import (
	"slices"
	"strings"
	"testing"
)

// frame is the frame the cases turn: the lens view, its source shown; fm,
// a slot of one tab; and the panel, its terminals shell, 3 and 4, as a
// close of 2 and a new leave them, 4 shown; hidden, or not.
func frame(hidden bool) []Slot {
	return []Slot{
		{Name: "view", Tabs: []string{"source", "outline", "render",
			"diff"}},
		{Name: "fm", Tabs: []string{"fm"}},
		{Name: "shell", Tabs: []string{"shell", "3", "4"}, Shown: 2,
			Hidden: hidden},
	}
}

// TestTurn: each word, as l and a tab line's press send it, means one
// thing to every host, which this table says once: a case is the frame,
// its panel hidden or not, the slot asked, the words, and what is asked
// of the host, or the error said.
func TestTurn(t *testing.T) {
	show := func(slot string, tab int) Act {
		return Act{Do: Show, Slot: slot, Tab: tab}
	}
	run := func(words ...string) Act { return Act{Do: New, Run: words} }
	for _, c := range []struct {
		hidden    bool
		at, words string
		want      Act
		said      string
	}{
		{false, "view", "next", show("view", 1), ""},
		{false, "view", "prev", show("view", 3), ""},
		{false, "shell", "next", show("shell", 0), ""},
		{false, "view", "3", show("view", 2), ""},
		{false, "shell", "2", show("shell", 1), ""}, // a place
		{false, "shell", "3", show("shell", 2), ""},
		{false, "view", "9", Act{}, "no tab 9 of view: it has 4"},
		{false, "shell", "outline", show("view", 1), ""},
		{false, "", "fm", show("fm", 0), ""},
		{false, "view", "nothing", Act{}, "no tab nothing"},
		{false, "view", "close", Act{Do: Close, Slot: "view"}, ""},
		{false, "fm", "close", Act{},
			"fm keeps its last tab: it is the slot"},
		{false, "fm", "next", Act{}, ""},
		{false, "", "next", Act{},
			"not in a frame's slot: name the tab"},
		{false, "view", "new", run(), ""},
		{false, "view", "new make test", run("make", "test"), ""},
		{true, "shell", "next", show("shell", 2), ""},
		{true, "shell", "close", show("shell", 2), ""},
		{true, "shell", "1", show("shell", 0), ""},
	} {
		words := strings.Fields(c.words)
		got, err := Turn(frame(c.hidden), c.at, words)
		said := ""
		if err != nil {
			said = err.Error()
		}
		if got.Do != c.want.Do || got.Slot != c.want.Slot ||
			got.Tab != c.want.Tab ||
			!slices.Equal(got.Run, c.want.Run) || said != c.said {
			t.Errorf("%q in %s, hidden %v: %+v %q, want %+v %q",
				c.words, c.at, c.hidden, got, said, c.want,
				c.said)
		}
	}
}
