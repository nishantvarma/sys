package picture

import (
	"strings"
	"testing"

	"sys/outline"
)

// pair is the picture the tests draw: two lanes, a box in each, one above
// the other, and a message down between them that carries a word.
const pair = "t\nup\n  a\n    → b  go\nlow\n  b\n"

// picture reads src as a note, as fm reads one, and lays it out.
func picture(t *testing.T, src string) *Picture {
	t.Helper()
	f, err := New(outline.Text(strings.NewReader(src), "note"))
	if err != nil {
		t.Fatal(err)
	}
	return f
}

// A message that names no box, two boxes, none above it, or a box it
// holds, is an error at its own line, which a click opens.
func TestBroken(t *testing.T) {
	for _, c := range [][2]string{
		{"t\na\n  → b\n", "note:3: → b: no box by that name"},
		{"t\na\n  b\nc\n  b\n  → b\n",
			"note:6: → b: 2 boxes by that name"},
		{"t\n→ a\na\n", "note:2: → a: under no box"},
		{"t\na\n  b\n    → a\n",
			"note:4: → a: one box holds the other"},
	} {
		_, err := New(outline.Text(strings.NewReader(c[0]), "note"))
		if err == nil || err.Error() != c[1] {
			t.Errorf("%q: %v, want %s", c[0], err, c[1])
		}
	}
}

// pair in runes: each box bordered round, the lanes' labels at the left,
// level with the first line in them, the message ┬ where it leaves and ▼
// where it arrives, and its word beside it, a cell clear, at the right,
// as the page says one message's word.
func TestRunes(t *testing.T) {
	got := strings.Join(picture(t, pair).Runes(), "\n")
	want := strings.Join([]string{
		"",
		"      ╭───╮",
		"up    │ a │",
		"      ╰─┬─╯",
		"        │",
		"        │",
		"        │ go",
		"        │",
		"        ▼",
		"      ╭───╮",
		"low   │ b │",
		"      ╰───╯",
		"",
	}, "\n")
	if got != want {
		t.Errorf("runes:\n%s\nwant:\n%s", got, want)
	}
}

// Two messages that leave a box by one cell, down to boxes one under
// another, keep the junction there: the second draws no line over the
// first's ┬.
func TestJunction(t *testing.T) {
	f := picture(t, "t\nup\n  a\n    → b\n    → c\nmid\n  b\nlow\n  c\n")
	if pic := strings.Join(f.Runes(), "\n"); !strings.Contains(pic,
		"╰─┬─╯") {
		t.Errorf("the junction is lost:\n%s", pic)
	}
}

// A word beside a line near an edge is in the picture, not cut by it: past
// the right edge the picture widens; past the left the whole of it moves
// right, and its lanes' labels stay at the left.
func TestFit(t *testing.T) {
	for _, c := range [][2]string{
		{"t\nup\n  a\n    → b  a long word\nlow\n  b\n",
			"a long word"},
		{"t\nup\n  a  a wide box note\n" +
			"    → b  first long word\n    → b  go\n" +
			"low\n  b  a wide box note\n", "first long word"},
	} {
		rows := picture(t, c[0]).Runes()
		if pic := strings.Join(rows, "\n"); !strings.Contains(pic,
			c[1]) {
			t.Errorf("%q is cut:\n%s", c[1], pic)
		}
		if !strings.HasPrefix(rows[2], "up ") {
			t.Errorf("the lane's label moved: %q", rows[2])
		}
	}
}

// A cell names what it falls on: the innermost box, its border too; a
// lane, on its band; a message, on its word; nothing on its line, in the
// column of labels, or off the picture.
func TestAt(t *testing.T) {
	f := picture(t, pair)
	for _, c := range []struct {
		x, y int
		want string
	}{
		{8, 2, "note:3"},
		{6, 1, "note:3"},
		{5, 2, "note:2"},
		{10, 6, "note:4"},
		{11, 6, "note:4"},
		{8, 5, ""},
		{0, 2, ""},
		{40, 40, ""},
	} {
		if got := f.At(c.x, c.y); got != c.want {
			t.Errorf("%d,%d: %q, want %q", c.x, c.y, got, c.want)
		}
	}
}

// pair on a page, whole: each box a card in a group that carries its
// item's address, the message's line and head in the accent, and its word
// in a group of its own.
func TestSVG(t *testing.T) {
	want := `<svg class="picture" viewBox="0 0 104 234" width="104"` +
		` height="234" role="img" aria-label="t">
<text class="lane" x="0" y="45" text-anchor="start">up</text>
<g data-id="note:2">
<rect class="ground" x="44" y="9" width="48" height="72" rx="6"/>
<g data-id="note:3">
<rect class="raised" x="52" y="27" width="32" height="36" rx="6"/>
<text class="label" x="68" y="45" text-anchor="middle">a</text>
</g>
</g>
<text class="lane" x="0" y="189" text-anchor="start">low</text>
<g data-id="note:5">
<rect class="ground" x="44" y="153" width="48" height="72" rx="6"/>
<g data-id="note:6">
<rect class="raised" x="52" y="171" width="32" height="36" rx="6"/>
<text class="label" x="68" y="189" text-anchor="middle">b</text>
</g>
</g>
<path class="msg" d="M68 63L68 165"/>
<path class="head" d="M68 171L64.5 164L71.5 164Z"/>
<g data-id="note:4">
<text class="said" x="76" y="117" text-anchor="start">go</text>
</g>
</svg>`
	if got := picture(t, pair).SVG(); got != want {
		t.Errorf("svg:\n%s\nwant:\n%s", got, want)
	}
}

// arch is a picture shaped as design/arch is: a lane of three boxes, the
// last holding one; messages across between neighbours and past one, its
// word longer than the room between any two; messages down, one word
// long enough to reach the next's line, the rest beside theirs on one
// row; and a bar.
const arch = `t
up
  a
    → b  across
    → c  a long word past b
    → d  a long word down
  b
    → d  down too
  c
    inner
      → d  from inside
low
  d
bar  a note
`

// Each word stands alone: on no box, no line and no other word, in runes
// as on the page, which say from one layout; and a cell of a word, one
// over a box the message passes too, names the message.
func TestWords(t *testing.T) {
	f := picture(t, arch)
	taken := f.taken()
	seen := map[[2]int]string{}
	for _, m := range f.arrows {
		x0, n := m.span()
		y := int(m.said[1])
		for x := x0; x < x0+n; x++ {
			if who, ok := seen[[2]int{x, y}]; ok {
				t.Errorf("%q and %q stand on %d,%d", who,
					m.label, x, y)
			}
			seen[[2]int{x, y}] = m.label
			if y < 0 || y >= len(taken) || x < 0 ||
				x >= len(taken[y]) || taken[y][x] {
				t.Errorf("%q stands on what is drawn,"+
					" at %d,%d", m.label, x, y)
			}
		}
		if got := f.At(x0, y); got != m.at {
			t.Errorf("a click on %q names %q, not %q", m.label,
				got, m.at)
		}
	}
	if t.Failed() {
		t.Log("\n" + strings.Join(f.Runes(), "\n"))
	}
}
