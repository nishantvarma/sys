package view

import (
	"regexp"
	"strconv"
	"strings"
	"testing"

	"sys/outline"
	"sys/tty"
)

// looks is how the tests' program draws its kinds: a folder in the
// accent, / after its name, a card saying folder.
var looks = map[string]Look{"dir": {Color: tty.Accent, Mark: "/",
	Word: "folder"}}

// sgr is a colour said in the terminal's words, for reading a screen
// without its colours (plain).
var sgr = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// plain is what a drawing says, without its colours.
func plain(s string) string { return sgr.ReplaceAllString(s, "") }

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
	for _, want := range []string{"/a / b", "b/", "c", "hi", "1/2"} {
		if !strings.Contains(plain(out), want) {
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
	if v.lay != 0 {
		t.Errorf("t in a figure turned the layout to %d", v.lay)
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

// A cell names the node drawn there: on the head, the last part of the
// path the cursor's and each part before it its folder; as a list, a
// row's anywhere on its line; as thumbs, a card's, not its margin or the
// gap after it; as a figure, its part or its message's word; nothing in
// the chrome, below the rows, or off the view.
func TestAt(t *testing.T) {
	_, list := region(tree(), 40, 12)
	thumbs := New(nil, "", looks)
	thumbs.In(0, 0, 60, 12)
	thumbs.last, thumbs.lay = tree(), 1
	thumbs.drawn("")
	_, fig := region(pair(), 20, 22)
	for _, c := range []struct {
		v    *Term
		x, y int
		want string
	}{
		{list, 0, 0, "/a"},
		{list, 3, 0, "/a"},
		{list, 6, 0, "/a/b"},
		{list, 3, 3, ""},
		{list, 3, 5, "/a/b"},
		{list, 39, 6, "/a/c"},
		{list, 3, 7, ""},
		{list, 3, 30, ""},
		{thumbs, 0, 5, ""},
		{thumbs, 1, 5, "/a/b"},
		{thumbs, 22, 6, "/a/b"},
		{thumbs, 23, 6, ""},
		{thumbs, 25, 5, "/a/c"},
		{fig, 0, 0, "n"},
		{fig, 9, 7, "n:3"},
		{fig, 11, 11, "n:4"},
		{fig, 9, 10, ""},
	} {
		if got := c.v.At(c.x, c.y); got != c.want {
			t.Errorf("%d,%d: %q, want %q", c.x, c.y, got, c.want)
		}
	}
}

// The two layouts of tree, each drawn in a region 40 by 12 and read
// without its colours: the head says where the cursor is, from home, or,
// as thumbs, the folder the cards are, and the keys; a rule; the layouts,
// the one in use lit; then the folder, its rows numbered; or thumbs, a
// card a thing, its kind's word under its name. The cursor's row, or
// card, is on the bar, with the accent's edge.
func TestLayouts(t *testing.T) {
	top := []string{
		" b                            t  layout",
		strings.Repeat("─", 40),
		"",
		" folder  thumbs",
		"",
	}
	rows := []string{"▌ 1  b/", "  2  c", "", "", "", ""}
	cards := []string{" ▌b/", "  folder", "", "  c", "  file", ""}
	home := append([]string{" ~                            t  layout"},
		top[1:5]...)
	for lay, want := range [][]string{
		append(top[:5:5], rows...),
		append(home, cards...),
	} {
		v := New(nil, "/a", looks)
		v.In(0, 0, 40, 12)
		v.last, v.lay = tree(), lay
		v.drawn("")
		var got []string
		for _, l := range v.drew {
			text := strings.TrimRight(plain(l.text()), " ")
			got = append(got, text)
		}
		if g, w := strings.Join(got, "\n"), strings.Join(want,
			"\n"); g != w {
			t.Errorf("%s:\n%s\nwant:\n%s", layouts[lay], g, w)
		}
		lit := tty.Paint(v.layout(), tty.Accent, tty.Ground, true)
		if !strings.Contains(v.drew[3].text(), lit) {
			t.Errorf("%s: the layout in use is not lit",
				layouts[lay])
		}
		if !strings.Contains(v.drew[5].text(), "48;2;28;42;61") {
			t.Errorf("%s: the cursor is not on the bar",
				layouts[lay])
		}
	}
}

// A folder's facts stand in one column: from 46, past the longest name
// that has them while there is room, else left enough to fit, the name
// cut short before them; not at all where no name would fit beside them.
func TestFacts(t *testing.T) {
	long := strings.Repeat("n", 50)
	for _, c := range []struct {
		w     int
		name  string
		at    int
		whole bool
	}{
		{60, "c", 46, true},   // from column 46
		{40, "c", 31, true},   // left enough to fit
		{90, long, 57, true},  // past the longest name, which stays
		{60, long, 51, false}, // the name cut short before them
		{20, "c", -1, true},   // too narrow: the name alone
	} {
		v := New(nil, "/a", looks)
		v.In(0, 0, c.w, 12)
		v.last = tree()
		v.last.Kids[1].Text = c.name
		v.last.Kids[1].Tags = []string{"640   1h"}
		v.drawn("")
		got := plain(v.drew[6].text())
		at := strings.Index(got, "640   1h")
		if at >= 0 {
			at = tty.Width(got[:at]) // a column, not a byte
		}
		if at != c.at {
			t.Errorf("%d wide, a name of %d: facts at %d in %q",
				c.w, len(c.name), at, got)
		}
		if strings.Contains(got, c.name) != c.whole {
			t.Errorf("%d wide, a name of %d: whole is %v in %q",
				c.w, len(c.name), !c.whole, got)
		}
	}
}

// A host that is no terminal gets the screen as runs: as many lines as
// its region is high, the bottom line last, saying what the terminal's
// would, and the cursor's row on the bar.
func TestLines(t *testing.T) {
	v := New(nil, "/a", looks)
	v.In(0, 0, 40, 12)
	v.Draw(tree(), "hi", nil)
	lines := v.Lines()
	if len(lines) != 12 {
		t.Fatalf("%d lines, want 12", len(lines))
	}
	say := func(rs []Run) string {
		var b strings.Builder
		for _, r := range rs {
			b.WriteString(r.Text)
		}
		return strings.TrimRight(b.String(), " ")
	}
	for i, want := range map[int]string{
		0:  " b" + strings.Repeat(" ", 28) + "t  layout",
		5:  "▌ 1  b/",
		11: " hi" + strings.Repeat(" ", 33) + "1/2",
	} {
		if got := say(lines[i]); got != want {
			t.Errorf("line %d: %q, want %q", i, got, want)
		}
	}
	if bg := lines[5][0].Ground; bg != tty.Bar {
		t.Errorf("the cursor's row is on %q, not the bar", bg)
	}
}

// vi's h j k l are the arrows: as thumbs, l and h go a card on and back;
// as rows, which have no left or right, the program hears h as left and l
// as right, to go out and in.
func TestVi(t *testing.T) {
	thumbs := New(nil, "", looks)
	thumbs.In(0, 0, 60, 12)
	thumbs.last, thumbs.lay = tree(), 1
	thumbs.drawn("")
	if ev, _ := thumbs.Handle(tty.Key{Name: "l"}); ev.Node != "/a/c" {
		t.Errorf("thumbs: l goes to %q, not /a/c", ev.Node)
	}
	thumbs.last.Kids[0].Cur, thumbs.last.Kids[1].Cur = false, true
	if ev, _ := thumbs.Handle(tty.Key{Name: "h"}); ev.Node != "/a/b" {
		t.Errorf("thumbs: h goes to %q, not /a/b", ev.Node)
	}
	_, list := region(tree(), 40, 12)
	for k, want := range map[string]string{"h": tty.KLeft,
		"l": tty.KRight} {
		if ev, _ := list.Handle(tty.Key{Name: k}); ev.Name != want {
			t.Errorf("rows: %s is heard as %q", k, ev.Name)
		}
	}
}

// A left press names what it falls on, at a cell the terminal counts from
// 1 and the view, in its region, from its own corner; a second soon after
// on the same cell opens it instead, and says nothing to the program; the
// wheel moves the cursor.
func TestMouse(t *testing.T) {
	t.Setenv("PATH", "") // no p to open anything with
	v := New(nil, "", looks)
	v.In(0, 2, 40, 12)
	v.last = tree()
	v.drawn("")
	press := tty.Key{Name: tty.KMouse, Button: tty.MLeft, X: 4, Y: 9}
	if ev, ok := v.mouse(press); !ok || ev.Node != "/a/c" {
		t.Errorf("a press on c's row names %q", ev.Node)
	}
	if ev, ok := v.mouse(press); ok {
		t.Errorf("a second press soon after names %q", ev.Node)
	}
	wheel := tty.Key{Name: tty.KMouse, Button: tty.MWheelDown}
	if ev, _ := v.mouse(wheel); ev.Node != "/a/c" {
		t.Errorf("the wheel down goes to %q, not /a/c", ev.Node)
	}
}
