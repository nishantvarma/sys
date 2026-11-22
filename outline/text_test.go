package outline

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// show is nodes as text again, two spaces a level, with their ids.
func show(b *strings.Builder, ns []*Node, depth int) {
	for _, n := range ns {
		b.WriteString(strings.Repeat("  ", depth) + n.Text)
		line := n.ID[strings.LastIndex(n.ID, ":"):]
		b.WriteString("   " + line + "\n")
		show(b, n.Kids, depth+1)
	}
}

// TestText: one note carried through, each rule in it once.
func TestText(t *testing.T) {
	in := `The plan
  1  a text file opens in place, every
     text file the same way
  2  tiles

  view     how it shows: zoomed, or as the tree
           from the top
  to the view    draw: the tree
                 and the status
- [ ] a step that
      wraps
  - a kid of the step
func f() {
	if x {
		return
	}
}
a paragraph runs
on, a line a node
`
	want := `The plan   :1
  1  a text file opens in place, every text file the same way   :2
  2  tiles   :4
  view     how it shows: zoomed, or as the tree from the top   :6
  to the view    draw: the tree and the status   :8
- [ ] a step that wraps   :10
  - a kid of the step   :12
func f() {   :13
  if x {   :14
    return   :15
  }   :16
}   :17
a paragraph runs   :18
on, a line a node   :19
`
	var b strings.Builder
	show(&b, Text(strings.NewReader(in), "note"), 0)
	if b.String() != want {
		t.Errorf("got\n%s\nwant\n%s", b.String(), want)
	}
}

// TestNotes: over every text file in sys, reading loses no word and moves
// none: the nodes, in order, say what the file says.
func TestNotes(t *testing.T) {
	files, nodes := 0, 0
	filepath.WalkDir("..", func(p string, d fs.DirEntry, err error) error {
		if d.IsDir() && (d.Name() == ".git" || d.Name() == "bin") {
			return filepath.SkipDir
		}
		b, err := os.ReadFile(p)
		if d.IsDir() || err != nil || bytes.IndexByte(b, 0) >= 0 {
			return nil
		}
		var said []string
		var walk func([]*Node)
		walk = func(ns []*Node) {
			for _, n := range ns {
				said = append(said, n.Text)
				nodes++
				walk(n.Kids)
			}
		}
		walk(Text(bytes.NewReader(b), p))
		got := strings.Fields(strings.Join(said, " "))
		want := strings.Fields(untab(string(b)))
		if strings.Join(got, " ") != strings.Join(want, " ") {
			t.Errorf("%s: the nodes say other than the file", p)
		}
		files++
		return nil
	})
	t.Logf("%d files, %d nodes", files, nodes)
}
