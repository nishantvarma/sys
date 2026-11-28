package view

import (
	"strings"

	"sys/outline"
	"sys/tty"
)

// A doc is a tree the view sets for the eye, as nroff sets a page: each
// part by its kind, prose filled to a measure, the names of rows
// side by side in a column, code on the panel, an address in the link's
// colour. It scrolls as a picture does; the part the program drew the
// cursor on is marked, and brought into sight when it moves. A kind it
// does not know is a head, its text, and its parts under it.

// isDoc is whether the tree drawn last is a doc.
func (v *Term) isDoc() bool {
	return v.last != nil && v.last.Kind == "doc"
}

// scrolls is whether the tree drawn last is read whole and scrolled, a
// picture or a doc, rather than rows a cursor moves among.
func (v *Term) scrolls() bool {
	return v.isPicture() || v.isDoc()
}

// typeset is the doc as its lines, w wide, each line's spot the part it
// sets, so that a click names that part: set once for a tree and a width,
// and kept, so that a scroll, a frame or the cursor moving sets nothing
// again (typed).
func (v *Term) typeset(w int) []line {
	if c := v.typed; c.tree == v.last && c.w == w {
		return c.lines
	}
	t := &setter{w: w}
	t.set(v.last.Kids, margin, false)
	v.typed = typed{v.last, w, t.out}
	return t.out
}

// typed is a doc as it was typeset last: the tree and the width it was set
// for, and its lines.
type typed struct {
	tree  *outline.Node
	w     int
	lines []line
}

// docked is the body as the doc, scrolled (scrolled), the part at the
// cursor brought into sight first when it is another than when it was
// drawn last, and its lines lit (lit).
func (v *Term) docked(w, room int) []line {
	ls, here := v.typeset(w), ""
	if c := v.cursor(); c != nil {
		here = c.ID
	}
	if here != v.seen {
		for i, l := range ls {
			if len(l.spots) > 0 && l.spots[0].id == here {
				v.top = block(ls, i-room/3, i-room+1)
				break
			}
		}
		v.seen = here
	}
	vis := v.scrolled(ls, w, room)
	for i, l := range vis {
		if len(l.spots) > 0 && l.spots[0].id == here {
			vis[i] = lit(l)
		}
	}
	return vis
}

// block is where a doc shown from line top starts, moved up to the first
// line of the block that line is in, the line after a blank one, so that a
// head is never cut from its rule nor a paragraph from its first line;
// but not above line least, past which what was to be in sight is not.
func block(ls []line, top, least int) int {
	top = max(0, top)
	for at := top; at > 0 && at > least; at-- {
		if len(ls[at-1].spots) == 0 {
			return at
		}
	}
	return top
}

// lit is a line of the part at the cursor, as a doc marks one: its edge
// the accent's ▌, and its ground the bar's, but where it is the panel's,
// a code's or a declaration's.
func lit(l line) line {
	k := &ink{w: l.ink.w}
	for i, r := range l.ink.runs {
		if r.Ground == "" {
			r.Ground = tty.Bar
		}
		if i == 0 {
			r.Text = "▌"
		}
		k.runs = append(k.runs, r)
	}
	return line{k, l.spots}
}

// The doc's measure: its left margin, and the most columns a line of its
// prose takes from it, as a page is set, however wide the view.
const (
	margin  = 2
	measure = 72
)

// setter sets a doc's parts as lines, w wide: the lines so far; the kind
// of the part set last and whether it held parts, which say the room
// before the next (gap); and the ground it sets text on, the panel's
// while it sets a declaration, whose text is code (set).
type setter struct {
	out  []line
	w    int
	last string
	held bool
	bg   string
}

// piece is some of a word, in its look: an address and the text around
// it are pieces of one word.
type piece struct {
	text, fg    string
	bold, under bool
}

// set sets parts, a declaration's when declared, their text from column
// in, and is whether it set any:
// a title bold; a paragraph filled; a code as written, on the panel; a
// row's name, bold, then its lead filled beside it, from a column past its
// fellows' names; an item's or a step's sign before its text; and any
// other kind, a section's or Go's, a head, in its look, bold, or its name
// bold when it declares one; a top head, and the title, ruled off; and
// under each its parts, two columns in, a row's from its column. A
// declaration is code, and stands on the panel, as a code does, across
// the measure, its lines past the first four columns in, as code goes on;
// its doc is prose, hung four columns under it, as a man page hangs a
// tag's, its first paragraph, the summary, bright, and its later ones
// soft, for the eye to skim.
func (t *setter) set(ns []*outline.Node, in int, declared bool) bool {
	col := column(ns) + 2 // where a row's text starts
	first := true
	for _, n := range ns {
		t.gap(n)
		parts, under := n.Kids, in+2
		if declares(n) {
			under = in + 4
		}
		switch n.Kind {
		case "title":
			t.fill(n, in, in, nil, tty.Fore, true)
			t.rule(n, in)
		case "para":
			fg := tty.Fore
			if declared && !first {
				fg = tty.Soft
			}
			first = false
			t.fill(n, in, in, nil, fg, false)
		case "code":
			t.code(n, in)
		case "row":
			said := lead(n)
			if said == nil {
				said = &outline.Node{ID: n.ID}
			}
			name := []piece{{text: printable(n.Text), fg: tty.Fore,
				bold: true}}
			t.fill(said, in, in+col, name, tty.Fore, false)
			parts, under = held(n), in+col
		case "item", "step":
			s, c := sign(n)
			mark := []piece{{text: s, fg: c}}
			t.fill(n, in, in+2, mark, tty.Fore, false)
		default:
			hang := in
			if declares(n) {
				hang, t.bg = in+4, tty.Panel
			}
			t.fill(n, in, hang, nil, look(n.Kind).Color, true)
			if n.Kind == "section" && in == margin {
				t.rule(n, in)
			}
			t.bg, t.last = "", "head"
		}
		t.last, t.held = n.Kind, t.set(parts, under, declares(n))
	}
	return len(ns) > 0
}

// gap is the room before part n: a blank line, but before the first and
// under a head; and leaves of a kind, rows, fields, a list's items and
// steps, stand together, as a table does, while a part that holds parts
// stands apart on both sides.
func (t *setter) gap(n *outline.Node) {
	kind := n.Kind
	block := kind == "title" || kind == "para" || kind == "code" ||
		kind == "section"
	list := func(k string) bool { return k == "item" || k == "step" }
	same := kind == t.last || list(kind) && list(t.last)
	holds := len(held(n)) > 0 || kind != "row" && lead(n) != nil
	if len(t.out) > 0 && t.last != "head" &&
		(block || !same || t.held || holds) {
		t.out = append(t.out, line{ink: t.start(t.w)})
	}
	t.last, t.held = kind, false
}

// fill sets a part's text, its words filled to the measure, in fg and
// bold as words says them: the first line from column in, and when it has
// a mark, a row's name or an item's sign, that first, its text then from
// column hang; and the rest of its lines from column hang.
func (t *setter) fill(n *outline.Node, in, hang int, mark []piece, fg string,
	bold bool) {
	right := t.right()
	k := t.start(in)
	for _, p := range mark {
		t.put(k, p)
	}
	if len(mark) > 0 {
		if k.w > hang { // a mark too wide for its column
			k.add(" ", "", "", false)
		}
		k.pad(hang, "")
	}
	fresh, last := true, piece{}
	for _, w := range words(n, fg, bold) {
		wide := 0
		for _, p := range w {
			wide += tty.Width(p.text)
		}
		if !fresh && k.w+1+wide > right {
			t.end(k, n.ID)
			k, fresh = t.start(hang), true
		}
		// a space between two words of an address is the address's
		if !fresh {
			in := last.under && w[0].under
			t.put(k, piece{" ", last.fg, false, in})
		}
		for _, p := range w {
			t.put(k, p)
		}
		fresh, last = false, w[len(w)-1]
	}
	t.end(k, n.ID)
}

// right is the column a doc's lines end at: the measure's, or the view's
// last but one, when it is narrower.
func (t *setter) right() int {
	return min(t.w-2, margin+measure)
}

// put puts a piece on a line, in its look, on the setter's ground.
func (t *setter) put(k *ink, p piece) {
	k.add(p.text, p.fg, t.bg, p.bold)
	if p.under {
		k.under()
	}
}

// rule rules a head off: a quiet line under it, as wide as its text, as a
// page sets a section's head apart.
func (t *setter) rule(n *outline.Node, in int) {
	wide := min(tty.Width(name(n)), measure, t.w-in-2)
	k := t.start(in)
	k.add(strings.Repeat("─", max(0, wide)), tty.Quiet, "", false)
	t.end(k, n.ID)
}

// code sets a code's lines as written, from column in, on the panel, all
// as wide as the widest.
func (t *setter) code(n *outline.Node, in int) {
	ls := strings.Split(n.Text, "\n")
	wide := 0
	for _, l := range ls {
		wide = max(wide, tty.Width(printable(l)))
	}
	for _, l := range ls {
		k := t.start(in)
		k.add(" "+printable(l), tty.Fore, tty.Panel, false)
		k.pad(in+wide+2, tty.Panel)
		t.end(k, n.ID)
	}
}

// start is a line from column in, its first column the edge, which the
// part at the cursor lights (lit); on a ground, from the column before in.
func (t *setter) start(in int) *ink {
	k := (&ink{}).add(" ", tty.Accent, "", false)
	if t.bg != "" {
		k.pad(in-1, "")
	}
	return k.pad(in, t.bg)
}

// end puts a line of part id among the lines, w wide, its spot the part;
// on a ground, to the column past the right edge.
func (t *setter) end(k *ink, id string) {
	k.clip(t.w)
	if t.bg != "" {
		k.pad(t.right()+1, t.bg)
	}
	k.pad(t.w, "")
	t.out = append(t.out, line{k, []spot{{x: 0, id: id}}})
}

// words is a part's text as its words, each its pieces in their looks:
// an address in the link's colour, underlined; the name a part declares
// bold, and the rest then not, as a man page sets a declaration; else the
// rest in fg, and bold when asked. Its runs, when it has them, else its
// text.
func words(n *outline.Node, fg string, bold bool) [][]piece {
	runs := n.Runs
	if len(runs) == 0 {
		runs = []*outline.Node{{Text: n.Text}}
	}
	if declares(n) {
		bold = false
	}
	var out [][]piece
	var cur []piece
	for _, r := range runs {
		p := piece{fg: fg, bold: bold}
		switch r.Kind {
		case "addr":
			p.fg, p.under = tty.Link, true
		case "name":
			p.bold = true
		}
		for i, f := range strings.Split(printable(r.Text), " ") {
			if i > 0 && len(cur) > 0 {
				out, cur = append(out, cur), nil
			}
			if f != "" {
				p.text = f
				cur = append(cur, p)
			}
		}
	}
	if len(cur) > 0 {
		out = append(out, cur)
	}
	return out
}

// word is the word a cell of a doc's part shows, from column x and row y
// counted from 0, as it was drawn last, for a click to say what it names:
// "" off a part's line, on the head, whose words are places, and off a
// doc, whose rows a click names whole (At).
func (v *Term) word(x, y int) string {
	if !v.isDoc() || y < 1 || y >= len(v.drew) ||
		len(v.drew[y].spots) == 0 {
		return ""
	}
	var text []rune
	for _, r := range v.drew[y].ink.runs {
		text = append(text, []rune(r.Text)...)
	}
	return outline.Word(text, x)
}
