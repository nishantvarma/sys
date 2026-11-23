package view

import (
	"fmt"
	"strings"
	"unicode"

	"sys/figure"
	"sys/outline"
	"sys/tty"
)

const (
	curBg = "\x1b[48;5;236m"
	selBg = "\x1b[48;5;239m"
)

// screen is the whole screen, w by h: the tree drawn last, laid out as the
// view lays it out, in h-1 lines, and the bottom line, with word on its
// left. Before any tree, only the bottom line.
func (v *Term) screen(word string, w, h int) ([]line, string) {
	bottom := tty.Clip(foot(word, v.state, v.words(), w), w)
	switch {
	case v.last == nil:
		return nil, bottom
	case v.figured():
		return v.pictured(v.last, w, h), bottom
	}
	head, rows := v.shown()
	if v.lay == "tiles" {
		return v.tiled(head, rows, w, h), bottom
	}
	return v.render(head, rows, w, h), bottom
}

// figured is whether the tree drawn last is a figure: a note whose kind
// says to draw it as its parts and the messages between them.
func (v *Term) figured() bool {
	return v.last != nil && v.last.Kind == "figure"
}

// render is a screen of rows as a list: the header, as many rows as fit,
// the cursor's kept in the middle when it can be, and a count of the rows
// that did not fit, h-1 lines in all.
func (v *Term) render(head *outline.Node, rows []row, w, h int) []line {
	lines := []line{v.header(head, w)}
	n := h - 2
	if len(rows) > n {
		n--
	}
	n = max(0, n)
	off := max(0, min(curRow(rows)-n/2, len(rows)-n)) // cursor row centred
	vis := rows[off:min(off+n, len(rows))]
	for i, r := range vis {
		lines = append(lines, line{tty.Clip(v.paint(off+i, r), w),
			[]spot{{0, r.n.ID}}})
	}
	extra := n - len(vis)
	if len(rows) > n {
		lines = append(lines, more(len(rows)-n, w))
		extra--
	}
	for i := 0; i < extra; i++ {
		lines = append(lines, line{})
	}
	return lines
}

// tiled is a screen of rows as tiles, a grid read left to right: the
// header, as many lines of tiles as fit, the cursor's line kept in the
// middle when it can be, and a count of the tiles that did not fit, as
// render is for a list.
func (v *Term) tiled(head *outline.Node, rows []row, w, h int) []line {
	cols, cw := v.grid(rows, w)
	var strips []line // a line of tiles each
	for i := 0; i < len(rows); i += cols {
		var s strings.Builder
		var spots []spot
		for j := i; j < min(i+cols, len(rows)); j++ {
			spots = append(spots, spot{tty.Width(s.String()),
				rows[j].n.ID})
			s.WriteString(v.tile(j, rows[j], cw))
		}
		strips = append(strips, line{tty.Clip(s.String(), w), spots})
	}
	lines := []line{v.header(head, w)}
	n := h - 2
	if len(strips) > n {
		n--
	}
	n = max(0, n)
	off := max(0, min(curRow(rows)/cols-n/2, len(strips)-n))
	vis := strips[off:min(off+n, len(strips))]
	lines = append(lines, vis...)
	extra := n - len(vis)
	if len(strips) > n {
		seen := min(len(rows), (off+len(vis))*cols) - off*cols
		lines = append(lines, more(len(rows)-seen, w))
		extra--
	}
	for i := 0; i < extra; i++ {
		lines = append(lines, line{})
	}
	return lines
}

// pictured is a screen of a figure, a note whose kind is figure: the
// header, then its items laid out as parts and messages (sys/figure) and
// drawn in runes, from the line the view has scrolled to, as many as fit,
// with a count of those below; a note that makes no figure says why. A
// cell of the picture names what the figure says stands there.
func (v *Term) pictured(n *outline.Node, w, h int) []line {
	f, err := figure.New(n.Kids)
	var pic []string
	if err != nil {
		pic = []string{tty.Dim(printable(err.Error()))}
	} else {
		for _, p := range f.Runes() {
			pic = append(pic, printable(p))
		}
	}
	rows := max(0, h-2)
	v.top = max(0, min(v.top, len(pic)-rows))
	vis := pic[v.top:min(len(pic), v.top+rows)]
	below := len(pic) - v.top - len(vis)
	if below > 0 && len(vis) > 0 {
		vis, below = vis[:len(vis)-1], below+1
	}
	lines := []line{v.header(n, w)}
	for i, p := range vis {
		l := line{text: tty.Clip(" "+p, w)}
		if f != nil {
			l.spots = parts(f, v.top+i, w)
		}
		lines = append(lines, l)
	}
	if below > 0 {
		lines = append(lines, more(below, w))
	}
	for len(lines) < h-1 {
		lines = append(lines, line{})
	}
	return lines
}

// parts is where the parts of f stand on its line y, drawn from a
// screen's second column, w wide: a spot each time what a cell names
// changes (figure.At).
func parts(f *figure.Figure, y, w int) []spot {
	var spots []spot
	last := ""
	for x := 0; x < w; x++ {
		if id := f.At(x-1, y); id != last {
			spots = append(spots, spot{x, id})
			last = id
		}
	}
	return spots
}

// more is the line that counts the n that did not fit, dim, cut to w.
func more(n, w int) line {
	return line{text: tty.Clip(tty.Dim(fmt.Sprintf(" +%d", n)), w)}
}

// header is the top line: where the view is, a path or an item's address,
// cut from the left when long, for its end names the place.
func (v *Term) header(n *outline.Node, w int) line {
	text := tty.Bold(tty.ClipLeft(printable(v.tilde(n.ID)), w))
	return line{text, []spot{{0, n.ID}}}
}

// paint is row r as line i: its number, its hint in the tree, and its name
// in its kind's look; outside the tree, where no hint says so, … after one
// shut with more inside.
func (v *Term) paint(i int, r row) string {
	look := v.look(r.n.Kind)
	num := tty.Yellow(fmt.Sprintf("%2d", i+1))
	name := look.Color(printable(r.n.Text) + look.Mark)
	lead := ""
	if r.lead != "" {
		lead = r.lead
	} else if !r.n.Open && len(r.n.Kids) > 0 {
		name += " …"
	}
	return fmt.Sprintf(" %s  %s%s%s", num, lead, hl(r.n), name)
}

// look is how the view draws a kind: as the program says, else plain.
func (v *Term) look(kind string) Look {
	l := v.looks[kind]
	if l.Color == nil {
		l.Color = tty.Plain
	}
	return l
}

// printable is s as a terminal shows it and nothing more: a tab as a
// space, and any other control character, which would move the cursor or
// change the colours, as ?. A name or a line of a note is drawn so.
func printable(s string) string {
	return strings.Map(func(r rune) rune {
		switch {
		case r == '\t':
			return ' '
		case unicode.IsControl(r):
			return '?'
		}
		return r
	}, s)
}

// hl is the background a row is on: the cursor's, the selection's, or none.
func hl(n *outline.Node) string {
	switch {
	case n.Cur:
		return curBg
	case n.Sel:
		return selBg
	default:
		return ""
	}
}

// tilde is a path with the home it starts in said as ~.
func (v *Term) tilde(p string) string {
	h := v.home
	if h != "" && (p == h || strings.HasPrefix(p, h+"/")) {
		return "~" + p[len(h):]
	}
	return p
}

// tile is row r as tile i: its number, then its label in its kind's look,
// cut or padded to cw so that the next tile lines up.
func (v *Term) tile(i int, r row, cw int) string {
	name := tty.Clip(v.label(r.n), cw)
	pad := strings.Repeat(" ", max(0, cw-tty.Width(name)))
	num := tty.Yellow(fmt.Sprintf("%2d", i+1))
	color := v.look(r.n.Kind).Color
	return " " + num + " " + hl(r.n) + color(name) + pad + " "
}

// grid is how many tiles fit across w, and how wide each tile's label is:
// the widest, up to 24, so that the tiles line up in columns. A tile is a
// space, its number, a space, its label and a space.
func (v *Term) grid(rows []row, w int) (cols, cw int) {
	for _, r := range rows {
		cw = max(cw, tty.Width(v.label(r.n)))
	}
	cw = min(cw, 24)
	return max(1, w/(cw+5)), cw
}

// label is how a node reads in a tile: its text and its kind's mark, and …
// when it is shut with more inside.
func (v *Term) label(n *outline.Node) string {
	s := printable(n.Text) + v.look(n.Kind).Mark
	if !n.Open && len(n.Kids) > 0 {
		s += " …"
	}
	return s
}

// foot is the bottom line, w wide: the program's word on the left, and on
// the right its state, then the view's, in the numbers' yellow. Short of
// room, the word is cut, for the state says what holds now.
func foot(word string, state, mine []string, w int) string {
	right := strings.Join(mine, " ")
	if len(state) > 0 {
		right = strings.Join(state, " ") + " · " + right
	}
	right = tty.Clip(printable(right), w)
	left := tty.Clip(printable(word), max(0, w-tty.Width(right)-1))
	pad := max(0, w-tty.Width(left)-tty.Width(right))
	return left + strings.Repeat(" ", pad) + tty.Yellow(right)
}

// words is what the view says of itself on the bottom line: its layout,
// and where the cursor is among what it shows, 12/34; a figure, which has
// no cursor, says only that it is one.
func (v *Term) words() []string {
	lay := v.lay
	switch {
	case v.figured():
		return []string{"figure"}
	case lay == "":
		lay = "folder"
	}
	if v.last == nil {
		return []string{lay}
	}
	_, rows := v.shown()
	if len(rows) == 0 {
		return []string{lay}
	}
	return []string{lay, fmt.Sprintf("%d/%d", curRow(rows)+1, len(rows))}
}
