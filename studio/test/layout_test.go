package test

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// placed is a slot as studio slots says it.
type placed struct {
	Name   string
	Weight int
	Run    string
	Kids   []placed
	Tabs   []placed
	Panel  bool
}

// spun is a tree of slots as a line: a split by its weight and its
// slots, the other way from its parent's, in brackets, a slot by its name
// and weight.
func spun(ps []placed) string {
	var out []string
	for _, p := range ps {
		if len(p.Kids) > 0 {
			split := "[" + spun(p.Kids) + "]" + itoa(p.Weight)
			out = append(out, strings.TrimSpace(p.Name+" ")+split)
			continue
		}
		out = append(out, p.Name+itoa(p.Weight))
	}
	return strings.Join(out, " ")
}

func itoa(n int) string {
	b, _ := json.Marshal(n)
	return string(b)
}

// TestLayout: the first slot that keeps a shell is the frame's panel,
// across the bottom, 38% high; the rest take room above it in the order
// they come, a golden spiral: the first 62% at the right, each next 62% of
// what is left, down and across in turn, the rest toward the bottom left;
// the lines under a slot are its tabs; and a share on a slot's line is an
// error, for the order gives the room.
func TestLayout(t *testing.T) {
	studio := built(t)
	slots := func(frame string) ([]placed, string) {
		note := filepath.Join(t.TempDir(), "frame")
		os.WriteFile(note, []byte(frame), 0o644)
		c := exec.Command(studio, "slots", note)
		out, err := c.CombinedOutput()
		if err != nil {
			return nil, string(out)
		}
		var ps []placed
		if err := json.Unmarshal(out, &ps); err != nil {
			t.Fatal(err)
		}
		return ps, ""
	}
	ps, why := slots("f\na  x\nb  y\nc\nd\n")
	if why != "" {
		t.Fatal(why)
	}
	if got := spun(ps); got != "[[b62 d38]38 a62]62 c38" ||
		!ps[1].Panel {
		t.Errorf("four slots: %s, the panel %v", got, ps[1].Panel)
	}
	ps, _ = slots("f\nview\n  one  x\n  two  y\nfm  fm\n")
	if got := spun(ps); got != "[fm38 view62]100" ||
		len(ps[0].Kids[1].Tabs) != 2 || ps[0].Kids[1].Run != "x" {
		t.Errorf("a slot of tabs: %s, its tabs %v", got,
			ps[0].Kids[1].Tabs)
	}
	if _, why := slots("f\na  40%  x\nb\n"); !strings.Contains(why,
		"not by a share") {
		t.Errorf("a share was taken: %q", why)
	}
}
