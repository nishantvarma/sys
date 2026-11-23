package view

import (
	"regexp"
	"strconv"
	"strings"
	"testing"

	"sys/outline"
	"sys/tty"
)

// looks is how the tests' program draws its kinds: a folder blue, with /
// after its name.
var looks = map[string]Look{"dir": {Color: tty.Blue, Mark: "/"}}

// tree is a folder a with a folder b, the cursor on it, and a file c.
func tree() *outline.Node {
	return &outline.Node{ID: "/a", Text: "a", Kids: []*outline.Node{
		{ID: "/a/b", Text: "b", Kind: "dir", Cur: true},
		{ID: "/a/c", Text: "c", Kind: "file"},
	}}
}

// pair is a note drawn as a figure: two lanes, a box in each, and a
// message down between them that carries a word.
func pair() *outline.Node {
	src := "t\nup\n  a\n    → b  go\nlow\n  b\n"
	return &outline.Node{ID: "n", Kind: "figure",
		Kids: outline.Text(strings.NewReader(src), "n")}
}

// region is a view of n alone in a region w by h at the terminal's top
// left, drawn: what it wrote, and the view.
func region(n *outline.Node, w, h int) (string, *Term) {
	v := New(nil, "", looks)
	v.In(0, 0, w, h)
	v.last = n
	return v.drawn("hi"), v
}

// placed is the rows a region's drawing puts at their corners.
var placed = regexp.MustCompile(`\x1b\[(\d+);(\d+)H`)

// A region is drawn in its place and nowhere else: each of its h rows
// starts at its own corner and is padded to its width, the header first
// and the bottom line last.
func TestRegion(t *testing.T) {
	v := New(nil, "", looks)
	v.In(10, 2, 20, 6)
	v.last = tree()
	out := v.drawn("hi")
	locs := placed.FindAllStringSubmatchIndex(out, -1)
	if len(locs) != 6 {
		t.Fatalf("%d rows placed, want 6", len(locs))
	}
	for i, l := range locs {
		row, col := out[l[2]:l[3]], out[l[4]:l[5]]
		if row != strconv.Itoa(3+i) || col != "11" {
			t.Errorf("row %d at %s;%s, want %d;11", i, row, col,
				3+i)
		}
		end := len(out)
		if i+1 < len(locs) {
			end = locs[i+1][0]
		}
		if w := tty.Width(out[l[1]:end]); w != 20 {
			t.Errorf("row %d is %d wide, want 20", i, w)
		}
	}
	for _, want := range []string{"/a", "b/", "c", "hi", "folder 1/2"} {
		if !strings.Contains(out, want) {
			t.Errorf("no %q in the region", want)
		}
	}
	if strings.Contains(out, tty.ClrEol) {
		t.Error("a region clears to the end of the terminal's line")
	}
}

// A region too small for the tree draws no more rows than it has: one
// row is the bottom line alone, and none draws nothing; a narrow one's
// count of rows unseen is cut to its width.
func TestSmall(t *testing.T) {
	for h := 0; h <= 3; h++ {
		out, _ := region(tree(), 20, h)
		if n := len(placed.FindAllString(out, -1)); n != h {
			t.Errorf("%d rows high: %d rows placed", h, n)
		}
	}
	out, _ := region(tree(), 2, 3)
	for _, l := range placed.Split(out, -1)[1:] {
		if w := tty.Width(l); w > 2 {
			t.Errorf("2 wide: a row is %d wide: %q", w, l)
		}
	}
}

// A name holding control characters is drawn with ? for each: it cannot
// move the cursor or clear the screen.
func TestPrintable(t *testing.T) {
	n := tree()
	n.Kids[1].Text = "x\x1b[2Jy"
	out, _ := region(n, 30, 6)
	if strings.Contains(out, "\x1b[2J") || !strings.Contains(out,
		"x?[2Jy") {
		t.Errorf("the name is drawn raw: %q", out)
	}
}

// A tree marks only what hides something: a folder, or a node with kids,
// shut gets ▸; an open one gets nothing, for its kids under it say so,
// and a leaf gets nothing.
func TestHints(t *testing.T) {
	x := &outline.Node{ID: "/a/x", Kind: "file"}
	root := &outline.Node{ID: "/", Kids: []*outline.Node{
		{ID: "/a", Kind: "dir", Open: true, Kids: []*outline.Node{x}},
		{ID: "/b", Kind: "dir"},
		{ID: "/c", Kind: "item", Kids: []*outline.Node{{ID: "/c:2"}}},
		{ID: "/d", Kind: "file"},
	}}
	want := map[string]string{"/a": "  ", "/a/x": "    ", "/b": shut,
		"/c": shut, "/d": "  "}
	for _, r := range flat(root, 0, "") {
		if r.lead != want[r.n.ID] {
			t.Errorf("%s: hint %q, want %q", r.n.ID, r.lead,
				want[r.n.ID])
		}
	}
}

// A key given to the view moves its cursor and hands the program only
// where it landed; t is the view's own, and the program hears nothing; a
// digit, and any key not the view's, is the program's.
func TestHandle(t *testing.T) {
	_, v := region(tree(), 20, 6)
	for _, c := range []struct {
		key        rune
		name, node string
		theirs     bool
	}{
		{'j', "", "/a/c", true},
		{'t', "", "", false},
		{'2', "2", "", true},
		{'x', "x", "", true},
	} {
		ev, ok := v.Handle(tty.Key{Rune: c.key})
		if ok != c.theirs || ev.Name != c.name || ev.Node != c.node {
			t.Errorf("%c: %+v %v, want %q %q %v", c.key, ev, ok,
				c.name, c.node, c.theirs)
		}
	}
}

// In a figure only scrolling is the view's: j a line down, g to the top,
// G to the bottom, where drawing stops it at the last page; t, a digit,
// and a move's left and right are the program's.
func TestScroll(t *testing.T) {
	_, v := region(pair(), 20, 6)
	for _, c := range []struct {
		key    tty.Key
		top    int
		theirs bool
	}{
		{tty.Key{Rune: 'j'}, 1, false},
		{tty.Key{Rune: 't'}, 1, true},
		{tty.Key{Rune: '3'}, 1, true},
		{tty.Key{Name: tty.KLeft}, 1, true},
		{tty.Key{Rune: 'g'}, 0, false},
		{tty.Key{Rune: 'G'}, 13, false},
	} {
		ev, ok := v.Handle(c.key)
		if ok != c.theirs || v.top != c.top ||
			ok && ev.Name != c.key.Token() {
			t.Errorf("%s: %+v %v top %d, want top %d",
				c.key.Token(), ev, ok, v.top, c.top)
		}
	}
	v.drawn("")
	if v.top != 9 {
		t.Errorf("drawn at top %d, want 9: 13 lines, 4 shown", v.top)
	}
	if v.lay != "" {
		t.Errorf("t in a figure turned the layout to %q", v.lay)
	}
}

// A figure scrolled down stays where it is while its tree is drawn again,
// and starts at its top when the tree is another.
func TestTop(t *testing.T) {
	_, v := region(pair(), 20, 6)
	v.Handle(tty.Key{Rune: 'j'})
	v.Draw(pair(), "", nil)
	if v.top != 1 {
		t.Errorf("drawn again: top %d, want 1", v.top)
	}
	other := pair()
	other.ID = "m"
	v.Draw(other, "", nil)
	if v.top != 0 {
		t.Errorf("another tree: top %d, want 0", v.top)
	}
}

// A cell names the node drawn there: as a list, a row's anywhere on its
// line, and the header the place; as tiles, each tile's; as a figure, its
// part or its message's word; nothing below the rows or off the view.
func TestAt(t *testing.T) {
	_, list := region(tree(), 20, 6)
	tiles := New(nil, "", looks)
	tiles.In(0, 0, 20, 6)
	tiles.last, tiles.lay = tree(), "tiles"
	tiles.drawn("")
	_, fig := region(pair(), 20, 16)
	for _, c := range []struct {
		v    *Term
		x, y int
		want string
	}{
		{list, 0, 0, "/a"},
		{list, 3, 1, "/a/b"},
		{list, 19, 2, "/a/c"},
		{list, 3, 3, ""},
		{list, 3, 5, ""},
		{list, 3, 9, ""},
		{tiles, 0, 1, "/a/b"},
		{tiles, 6, 1, "/a/b"},
		{tiles, 7, 1, "/a/c"},
		{fig, 0, 0, "n"},
		{fig, 9, 3, "n:3"},
		{fig, 11, 7, "n:4"},
		{fig, 9, 6, ""},
	} {
		if got := c.v.At(c.x, c.y); got != c.want {
			t.Errorf("%d,%d: %q, want %q", c.x, c.y, got, c.want)
		}
	}
}

// The three layouts of tree, each drawn in a region 24 by 6 and read
// without its colours: the folder, its rows numbered; the tree, the shut
// folder marked ▸; tiles, a grid of numbered names. The top line says
// where, the home as ~, and the cursor's row is lit.
func TestLayouts(t *testing.T) {
	sgr := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	for lay, want := range map[string][]string{
		"":      {"~", "  1  b/", "  2  c", "", ""},
		"tree":  {"~", "  1  ▸ b/", "  2    c", "", ""},
		"tiles": {"~", "  1 b/   2 c  ", "", "", ""},
	} {
		v := New(nil, "/a", looks)
		v.In(0, 0, 24, 6)
		v.last, v.lay = tree(), lay
		v.drawn("")
		var got []string
		for _, l := range v.drew {
			got = append(got, sgr.ReplaceAllString(l.text, ""))
		}
		if g, w := strings.Join(got, "\n"), strings.Join(want,
			"\n"); g != w {
			t.Errorf("%q:\n%s\nwant:\n%s", lay, g, w)
		}
		if !strings.Contains(v.drew[1].text, curBg) {
			t.Errorf("%q: the cursor's row is not lit", lay)
		}
	}
}
