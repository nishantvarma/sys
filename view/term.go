// Package view is the outline drawn on a terminal: fm's view, lifted out
// of fm so that a host, studio, draws it too. It lays a folder out itself,
// as its rows or as thumbs, cards read left to right, and it moves the
// cursor among what it shows, telling the program only where it landed
// (`design/view:/Who holds what/`). It draws in the palette, as the
// samples drew it: a head that says where the cursor is, the layouts over
// the rows, names the loudest thing, the cursor's row lit with a blue
// edge. It draws on the whole terminal, or in a region of it that a host
// gives it, and it knows which node each cell shows, for a click to name
// (At).
package view

import (
	"strings"

	"sys/outline"
	"sys/tty"
)

// Look is how the terminal draws a kind: its colour, from tty's palette;
// the mark after its name, as / after a folder's; the word a card says it
// by, folder; and the colour of what is said beside it, its facts. The
// program says how its kinds look; what it leaves out is the palette's
// foreground, and facts faint.
type Look struct {
	Color, Mark, Word, Said string
}

// Tab is one of the program's ways to see its thing, the switch line's in
// place of the view's layouts (Tabs): its name, a tally of what it shows,
// whether it is the one shown, and whether it waits to be built.
type Tab struct {
	Name, Tally string
	On, Waits   bool
}

// Term is the view: the tree drawn last, the program's word and state,
// its own layout, and the lines it drew; in a region, the rows and
// columns it is given.
type Term struct {
	t      *tty.Term
	home   string          // the folder the top line says as ~
	looks  map[string]Look // how the program's kinds are drawn
	last   *outline.Node   // the tree drawn last
	status string
	state  []string // the program's state, drawn last
	lay    int      // the layout in use, of layouts
	region bool     // in x, y, w, h, not the whole terminal
	x, y   int
	w, h   int
	top    int         // a figure's first line shown
	drew   []line      // the lines drawn last, all but the bottom one
	keys   [][2]string // the program's keys the head shows (Chips)
	tabs   []Tab       // the program's tabs (Tabs), nil for the layouts
}

// line is a line of the screen: its runs of text, and where on it each
// node is drawn, for At.
type line struct {
	ink   *ink
	spots []spot
}

// text is the line in the terminal's words.
func (l line) text() string { return l.ink.String() }

// spot is where a node starts on a line: it is drawn from column x up to
// the next spot, or to the line's end.
type spot struct {
	x  int
	id string
}

// New is the view on the whole of t, as fm draws itself: home is the
// folder its top line says as ~, and looks is how it draws each kind.
func New(t *tty.Term, home string, looks map[string]Look) *Term {
	return &Term{t: t, home: home, looks: looks}
}

// In puts the view in a region of the terminal: w by h, its top left at
// column x and row y, counted from 0. A host gives each view its own.
func (v *Term) In(x, y, w, h int) {
	v.region, v.x, v.y, v.w, v.h = true, x, y, w, h
}

// Draw shows the tree, the program's word and its state. A figure
// scrolled down starts at its top again when the tree is another.
func (v *Term) Draw(n *outline.Node, status string, state []string) {
	if v.last == nil || n.ID != v.last.ID {
		v.top = 0
	}
	v.last, v.status, v.state = n, status, state
	v.show(status)
}

// Chips are the keys the head shows at its right, each a key and what it
// does, the program's before the view's own; a range, 1..4, is a key.
func (v *Term) Chips(keys ...[2]string) {
	v.keys = keys
}

// Tabs are the program's ways to see its thing, which the switch line
// shows in place of the view's layouts; nil gives the layouts back.
func (v *Term) Tabs(tabs []Tab) {
	v.tabs = tabs
}

// Size is how much the view shows at once: the terminal, or its region.
func (v *Term) Size() (int, int) {
	if v.region {
		return v.w, v.h
	}
	return v.t.Size()
}

// Copy puts text on X's clipboard.
func (v *Term) Copy(text string) error {
	return tty.Put(tty.Clipboard, text)
}

// At is the node a cell of the view shows, column x and row y counted
// from the view's top left at 0, 0, as it was drawn last: a row's, a
// tile's, a figure's part or message, or, on the top line, the place the
// view is; "" where it shows none. A click names what it falls on.
func (v *Term) At(x, y int) string {
	if y < 0 || y >= len(v.drew) {
		return ""
	}
	id := ""
	for _, s := range v.drew[y].spots {
		if s.x <= x {
			id = s.id
		}
	}
	return id
}

// show draws the screen, with word on the left of the bottom line, and
// keeps the lines it drew, at most the h-1 that fit above the bottom line.
// On the whole terminal it is a frame of tty's, which writes the rows that
// changed, the cursor hidden; in a region, drawn. A view with no terminal,
// as a test makes, draws nothing.
func (v *Term) show(word string) {
	if v.t == nil {
		return
	}
	if v.region {
		v.t.Write(v.drawn(word))
		v.t.Flush()
		return
	}
	var rows []string
	for _, l := range v.page(word) {
		rows = append(rows, l.text())
	}
	v.t.Frame(rows, "")
}

// page is the screen as its h lines, word on the left of the bottom one,
// the last; it keeps the lines above the bottom, at most h-1, for At.
func (v *Term) page(word string) []line {
	w, h := v.Size()
	lines, bottom := v.screen(word, w, h)
	v.drew = lines[:min(len(lines), max(0, h-1))]
	if h < 1 {
		return nil
	}
	out := append([]line(nil), v.drew...)
	for len(out) < h-1 {
		out = append(out, blank(w))
	}
	return append(out, bottom)
}

// Lines is the screen as Draw left it, for a host that is no terminal and
// draws it itself, as gio does: a line a row, h of them for the region it
// is in (In), the bottom line last, each the runs it is drawn in. Where a
// click falls is At's, as on the terminal.
func (v *Term) Lines() [][]Run {
	var out [][]Run
	for _, l := range v.page(v.status) {
		out = append(out, l.ink.runs)
	}
	return out
}

// drawn is what show writes in a region: each row put at its own corner
// and padded to the region's width, so that nothing beside the region is
// touched, the bottom line last; a region with no rows draws nothing.
func (v *Term) drawn(word string) string {
	w, _ := v.Size()
	var b strings.Builder
	for i, l := range v.page(word) {
		b.WriteString(tty.Goto(v.y+i+1, v.x+1) + l.text() + tty.Sgr0)
		b.WriteString(strings.Repeat(" ", max(0, w-l.ink.w)))
	}
	return b.String()
}
