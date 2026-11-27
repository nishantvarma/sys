package picture

import (
	"math"
	"strings"
)

// cells is a picture's runes, a row a line, blank to start.
type cells [][]rune

// grid is w by h blank cells.
func grid(w, h int) cells {
	g := make(cells, h)
	for i := range g {
		g[i] = []rune(strings.Repeat(" ", w))
	}
	return g
}

// put sets the cell at x, y, if the picture has one there.
func (g cells) put(x, y int, r rune) {
	if y >= 0 && y < len(g) && x >= 0 && x < len(g[y]) {
		g[y][x] = r
	}
}

// line puts a message's rune at x, y where the cell is blank, and ┼ where
// it crosses another's line; a box's border and words it leaves as they
// are, so a line that must cross a box passes behind it.
func (g cells) line(x, y int, r rune) {
	if y < 0 || y >= len(g) || x < 0 || x >= len(g[y]) {
		return
	}
	switch c := g[y][x]; {
	case c == ' ':
		g[y][x] = r
	case c == '│' && r == '─', c == '─' && r == '│':
		g[y][x] = '┼'
	}
}

// write puts s's runes from x along row y.
func (g cells) write(x, y int, s string) {
	for i, r := range []rune(s) {
		g.put(x+i, y, r)
	}
}

// centre writes s centred on mid, in cells, along row y.
func (g cells) centre(mid float64, y int, s string) {
	g.write(int(math.Round(mid-float64(runes(s))/2)), y, s)
}

// border draws a box's border, round, in its outer cells.
func (g cells) border(x, y, w, h int) {
	if w < 2 || h < 2 {
		return
	}
	for i := x + 1; i < x+w-1; i++ {
		g.put(i, y, '─')
		g.put(i, y+h-1, '─')
	}
	for j := y + 1; j < y+h-1; j++ {
		g.put(x, j, '│')
		g.put(x+w-1, j, '│')
	}
	g.put(x, y, '╭')
	g.put(x+w-1, y, '╮')
	g.put(x, y+h-1, '╰')
	g.put(x+w-1, y+h-1, '╯')
}

// leave marks where a line leaves a box: a junction in its border, one
// way or the other, kept when a second line leaves by the same cell; or
// the line itself where no border is, as a band's.
func (g cells) leave(x, y int, fwd, back, line rune, forward bool) {
	if y < 0 || y >= len(g) || x < 0 || x >= len(g[y]) {
		return
	}
	switch g[y][x] {
	case '─', '│', fwd, back:
		g[y][x] = pick(forward, fwd, back)
	default:
		g[y][x] = line
	}
}

// lines is the cells as text, each row's trailing blanks cut.
func (g cells) lines() []string {
	out := make([]string, len(g))
	for i, row := range g {
		out[i] = strings.TrimRight(string(row), " ")
	}
	return out
}

// turn is the corner where a line moving down (d 1) or up (d -1) goes on
// right (h 1) or left (h -1): it joins the side the line came from to the
// side it leaves by. Where a line moving across goes on down or up, the
// corner joins the other two sides, turn(-d, -h).
func turn(d, h int) rune {
	switch {
	case d > 0 && h > 0:
		return '└'
	case d > 0 && h < 0:
		return '┘'
	case d < 0 && h > 0:
		return '┌'
	default:
		return '┐'
	}
}

// sign is n's direction: 1, -1, or 0.
func sign(n int) int {
	switch {
	case n > 0:
		return 1
	case n < 0:
		return -1
	}
	return 0
}

// pick is a when yes, else b.
func pick(yes bool, a, b rune) rune {
	if yes {
		return a
	}
	return b
}

// runes draws b and what it holds: its border, then, when it holds none,
// its label and note in it, centred, from its top, and when it holds
// boxes, its title at its top and its boxes in it. A bar, a lane alone,
// has its note in it and its label at the left.
func (b *box) runes(g cells, depth int) {
	g.border(b.x, b.y, b.w, b.h)
	row := b.y + 1
	if len(b.kids) == 0 {
		mid := float64(b.x) + float64(b.w)/2
		if depth > 0 {
			g.centre(mid, row, b.label)
			row++
		}
		for i, t := range b.lines {
			g.centre(mid, row+i, t)
		}
		return
	}
	g.write(b.x+2, row, b.title())
	for _, k := range b.kids {
		k.runes(g, depth+1)
	}
}

// runes draws m's line from its tail, in the border it leaves, to its
// head, beside the border it arrives at: the cells its points are in.
func (m *arrow) runes(g cells) {
	x0, y0 := int(m.tail[0]), int(m.tail[1])
	x1, y1 := int(m.head[0]), int(m.head[1])
	if y0 == y1 {
		d := sign(x1 - x0)
		g.leave(x0, y0, '├', '┤', '─', d > 0)
		for x := x0 + d; (x1-d-x)*d > 0; x += d {
			g.line(x, y0, '─')
		}
		g.put(x1-d, y0, pick(d > 0, '▶', '◀'))
		return
	}
	d := sign(y1 - y0)
	g.leave(x0, y0, '┬', '┴', '│', d > 0)
	bend := y1
	if x0 != x1 {
		bend = y0 + (y1-y0)/2
	}
	for y := y0 + d; (bend-y)*d > 0 && (y1-d-y)*d > 0; y += d {
		g.line(x0, y, '│')
	}
	if x0 != x1 {
		h := sign(x1 - x0)
		g.line(x0, bend, turn(d, h))
		for x := x0 + h; (x1-x)*h > 0; x += h {
			g.line(x, bend, '─')
		}
		g.line(x1, bend, turn(-d, -h))
		for y := bend + d; (y1-d-y)*d > 0; y += d {
			g.line(x1, y, '│')
		}
	}
	g.put(x1, y1-d, pick(d > 0, '▼', '▲'))
}

// runesSaid writes what m carries where the page says it (see span):
// beside a line going down, a cell clear of it, or over or under one
// going across.
func (m *arrow) runesSaid(g cells) {
	if x0, n := m.span(); n > 0 {
		g.write(x0, int(m.said[1]), m.label)
	}
}

// Runes is the picture in a terminal's cells, a line a row, from the same
// layout the page draws. A box is drawn round, its border in its outer
// cells, as the page's runs through their middle; a lane stands as a
// band, its label and note in the column at left, and a lane that holds
// no boxes is a bar, bordered, its note in it. A message is a line of │ or
// ─ from the side of one box to the other's: ┬ ┴ ├ ┤ where it leaves a
// border, ▼ ▲ ▶ ◀ where it arrives, bent once at the middle where the two
// boxes do not line up; what it carries stands beside it.
func (f *Picture) Runes() []string {
	g := f.drawing()
	for _, m := range f.arrows {
		m.runesSaid(g)
	}
	return g.lines()
}

// drawing is the picture's boxes and lines in cells, its words aside,
// which go over them (Runes), and clear of them (say).
func (f *Picture) drawing() cells {
	g := grid(f.w, f.h)
	for _, l := range f.lanes {
		row := l.y + 1
		if len(l.kids) > 0 {
			row++
		}
		g.write(0, row, l.label)
		if len(l.kids) == 0 {
			l.runes(g, 0)
			continue
		}
		for i, t := range l.lines {
			g.write(0, row+1+i, t)
		}
		for _, k := range l.kids {
			k.runes(g, 1)
		}
	}
	for _, m := range f.arrows {
		m.runes(g)
	}
	return g
}
