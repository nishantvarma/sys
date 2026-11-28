package view

import (
	"strings"

	"sys/outline"
	"sys/tty"
)

// A doc is a tree the view sets for the eye, as nroff sets a page: each
// part by its kind, prose filled to the view's width, the names of rows
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

// here is the id of the part the program drew the cursor on, the one a
// doc marks: the row it stands on (rowOf); "" for none.
func (v *Term) here() string {
	if c := v.marked(); c != nil {
		return v.rowOf(c).ID
	}
	return ""
}

// typeset is the doc as its lines, w wide, each line's spot the part it
// sets, so that a click names that part.
func (v *Term) typeset(w int) []line {
	t := &setter{w: w, at: v.here()}
	t.set(v.last.Kids, margin)
	return t.out
}

// docked is the body as the doc, scrolled (scrolled), the part at the
// cursor brought into sight first when it is another than when it was
// drawn last.
func (v *Term) docked(w, room int) []line {
	ls := v.typeset(w)
	if here := v.here(); here != v.seen {
		for i, l := range ls {
			if len(l.spots) > 0 && l.spots[0].id == here {
				v.top = i - room/3
				break
			}
		}
		v.seen = here
	}
	return v.scrolled(ls, w, room)
}

// The doc's measure: its left margin, and the most columns a line of its
// prose takes from it, as a page is set, however wide the view.
const (
	margin  = 2
	measure = 72
)

// setter sets a doc's parts as lines, w wide: the lines so far; the kind
// of the part set last and whether it held parts, which say the room
// before the next (gap); and the id of the part at the cursor, whose
// lines it marks.
type setter struct {
	out  []line
	w    int
	last string
	held bool
	at   string
}

// piece is some of a word, in its look: an address and the text around
// it are pieces of one word.
type piece struct {
	text, fg    string
	bold, under bool
}

// set sets parts, their text from column in, and is whether it set any:
// a title bold; a paragraph filled; a code as written, on the panel; a
// row's name, bold, then its lead filled beside it, from a column past its
// fellows' names; an item's or a step's sign before its text; and any
// other kind, a section's or Go's, a head, in its look, bold, or its name
// bold when it declares one; a top head, and the title, ruled off; and
// under each its parts, two columns in, a row's from its column.
func (t *setter) set(ns []*outline.Node, in int) bool {
	col := column(ns) + 2 // where a row's text starts
	any := false
	for _, n := range ns {
		if outline.IsRun(n) {
			continue
		}
		any = true
		t.gap(n.Kind)
		parts, under := n.Kids, in+2
		switch n.Kind {
		case "title":
			t.fill(n, in, in, nil, tty.Fore, true)
			t.rule(n, in)
		case "para":
			t.fill(n, in, in, nil, tty.Fore, false)
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
			t.fill(n, in, in, nil, look(n.Kind).Color, true)
			if n.Kind == "section" && in == margin {
				t.rule(n, in)
			}
			t.last = "head"
		}
		t.last, t.held = n.Kind, t.set(parts, under)
	}
	return any
}

// gap is the room before a part of kind: a blank line, but before the
// first and under a head; and leaves of a kind, rows, fields, a list's
// items and steps, stand together, as a table does.
func (t *setter) gap(kind string) {
	block := kind == "title" || kind == "para" || kind == "code" ||
		kind == "section"
	list := func(k string) bool { return k == "item" || k == "step" }
	same := kind == t.last || list(kind) && list(t.last)
	if len(t.out) > 0 && t.last != "head" && (block || !same || t.held) {
		t.out = append(t.out, blank(t.w))
	}
	t.last, t.held = kind, false
}

// fill sets a part's text, its words filled to the measure, the first
// line from column in, after mark, and the rest from column hang, in fg
// and bold as words says them.
func (t *setter) fill(n *outline.Node, in, hang int, mark []piece, fg string,
	bold bool) {
	right := min(t.w-2, margin+measure)
	k := t.start(n.ID, in)
	for _, p := range mark {
		t.put(k, p, n.ID)
	}
	if k.w > hang { // a mark too wide for its column
		k.add(" ", "", t.ground(n.ID), false)
	}
	k.pad(hang, t.ground(n.ID))
	fresh, last := true, piece{}
	for _, w := range words(n, fg, bold) {
		wide := 0
		for _, p := range w {
			wide += tty.Width(p.text)
		}
		if !fresh && k.w+1+wide > right {
			t.end(k, n.ID)
			k, fresh = t.start(n.ID, hang), true
		}
		// a space between two words of an address is the address's
		if !fresh {
			in := last.under && w[0].under
			t.put(k, piece{" ", last.fg, false, in}, n.ID)
		}
		for _, p := range w {
			t.put(k, p, n.ID)
		}
		fresh, last = false, w[len(w)-1]
	}
	t.end(k, n.ID)
}

// put puts a piece on a line of part id, in its look.
func (t *setter) put(k *ink, p piece, id string) {
	k.add(p.text, p.fg, t.ground(id), p.bold)
	if p.under {
		k.under()
	}
}

// rule rules a head off: a quiet line under it, as wide as its text, as a
// page sets a section's head apart.
func (t *setter) rule(n *outline.Node, in int) {
	wide := min(tty.Width(name(n)), measure, t.w-in-2)
	k := t.start(n.ID, in)
	k.add(strings.Repeat("─", max(0, wide)), tty.Quiet, t.ground(n.ID),
		false)
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
		k := t.start(n.ID, in)
		k.add(" "+printable(l), tty.Fore, tty.Panel, false)
		k.pad(in+wide+2, tty.Panel)
		t.end(k, n.ID)
	}
}

// start is a line of part id, from column in: its first column the
// cursor's edge when the part is the one at the cursor.
func (t *setter) start(id string, in int) *ink {
	edge := " "
	if id == t.at {
		edge = "▌"
	}
	return (&ink{}).add(edge, tty.Accent, t.ground(id), false).
		pad(in, t.ground(id))
}

// end puts a line of part id among the lines, w wide, its spot the part.
func (t *setter) end(k *ink, id string) {
	k.clip(t.w).pad(t.w, t.ground(id))
	t.out = append(t.out, line{k, []spot{{x: 0, id: id}}})
}

// ground is what a line of part id is drawn on: the bar's colour for the
// part at the cursor, else the view's own.
func (t *setter) ground(id string) string {
	if id == t.at {
		return tty.Bar
	}
	return ""
}

// words is a part's text as its words, each its pieces in their looks:
// an address in the link's colour, underlined; the name a part declares
// bold, and the rest then not, as a man page sets a declaration; else the
// rest in fg, and bold when asked. Its runs, when it has them, else its
// text.
func words(n *outline.Node, fg string, bold bool) [][]piece {
	runs := runsOf(n)
	if len(runs) == 0 {
		runs = []*outline.Node{{Text: n.Text}}
	}
	for _, r := range runs {
		if r.Kind == "name" {
			bold = false
		}
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

// Word is the word a cell of a doc's part shows, from column x and row y
// counted from 0, as it was drawn last, for a click to say what it names:
// "" off a part's line, on the head, whose words are places, and off a
// doc, whose rows a click names whole (At).
func (v *Term) Word(x, y int) string {
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
