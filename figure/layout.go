package figure

import (
	"fmt"
	"math"
)

// holds is whether b is x, or holds it at any depth.
func (b *box) holds(x *box) bool {
	if b == x {
		return true
	}
	for _, k := range b.kids {
		if k.holds(x) {
			return true
		}
	}
	return false
}

// title is what heads a box that holds boxes: its label, then its note.
func (b *box) title() string {
	if b.note == "" {
		return b.label
	}
	return b.label + "  " + b.note
}

// apart is the room between two neighbouring boxes, or lanes, where a
// message runs between them or what they hold. Across: room for its
// label, or its head at least, else the usual gap. Down: room for a
// label, else a line.
func (f *Figure) apart(a, b *box, across bool) int {
	g := 1
	if across {
		g = gapW
	}
	for _, m := range f.arrows {
		if a.holds(m.from) && b.holds(m.to) ||
			b.holds(m.from) && a.holds(m.to) {
			if across {
				g = max(g, runes(m.label)+4)
			} else {
				g = max(g, gapH)
			}
		}
	}
	return g
}

// size sets b's own size in cells, and its kids', before any is stretched
// to its neighbours. A box is as big as its label and its note; one that
// holds boxes, as big as they are with its title above them, set across
// at an even depth and down at an odd one. A lane is depth 0, so its boxes
// go across; its label and note stand in the column at its left, and a
// lane that holds no boxes is a bar, its note in it on one line.
func (b *box) size(depth int, f *Figure) {
	if len(b.kids) == 0 && depth == 0 {
		b.lines = wrap(b.note, math.MaxInt)
		b.w, b.h = widest(b.lines...)+4, 3
		return
	}
	if len(b.kids) == 0 {
		b.lines = wrap(b.note, noteW)
		b.w = widest(append(b.lines, b.label)...) + 4
		b.h = len(b.lines) + 3
		return
	}
	across := depth%2 == 0
	w, h := 0, 0
	for i, k := range b.kids {
		k.size(depth+1, f)
		if across {
			if i > 0 {
				w += f.apart(b.kids[i-1], k, true)
			}
			w, h = w+k.w, max(h, k.h)
		} else {
			if i > 0 {
				h += f.apart(b.kids[i-1], k, false)
			}
			w, h = max(w, k.w), h+k.h
		}
	}
	if depth == 0 {
		b.lines = wrap(b.note, laneW)
		b.w, b.h = w+2, max(h+2, len(b.lines)+3)
		return
	}
	b.w, b.h = max(w+2, runes(b.title())+4), h+3
}

// place puts b at x, y, w wide and h high, stretched from its own size,
// and its kids inside it, under its title: across, each wider by an even
// share of the room spare, and all as high as b holds; down, each as wide
// as b holds, and higher by a share.
func (b *box) place(x, y, w, h, depth int, f *Figure) {
	b.x, b.y, b.w, b.h = x, y, w, h
	if len(b.kids) == 0 {
		return
	}
	ix, iy, iw, ih := x+1, y+2, w-2, h-3
	if depth == 0 { // a lane has no title in it
		iy, ih = y+1, h-2
	}
	across := depth%2 == 0
	spare := iw
	if !across {
		spare = ih
	}
	for i, k := range b.kids {
		if i > 0 {
			spare -= f.apart(b.kids[i-1], k, across)
		}
		if across {
			spare -= k.w
		} else {
			spare -= k.h
		}
	}
	n := len(b.kids)
	at := ix
	if !across {
		at = iy
	}
	for i, k := range b.kids {
		if i > 0 {
			at += f.apart(b.kids[i-1], k, across)
		}
		share := spare / n
		if i < spare%n {
			share++
		}
		if across {
			k.place(at, iy, k.w+share, ih, depth+1, f)
			at += k.w
		} else {
			k.place(ix, at, iw, k.h+share, depth+1, f)
			at += k.h
		}
	}
}

// down lays the messages between a box and one below it, from the side of
// one that faces the other to the side that faces back: straight down
// where the two overlap, slanted between their near ends where they do
// not. The first half of them have their labels at their left, the rest
// at their right, so two that pass each other are said apart.
func down(a, b *box, ms []*arrow) {
	up, low := a, b
	if b.y < a.y {
		up, low = b, a
	}
	y0, y1 := float64(up.y+up.h)-0.5, float64(low.y)+0.5
	lo := float64(max(up.x, low.x)) + 1.5
	hi := float64(min(up.x+up.w, low.x+low.w)) - 1.5
	n := len(ms)
	for k, m := range ms {
		x0 := lo + (hi-lo)*float64(k+1)/float64(n+1)
		x1 := x0
		if hi < lo { // they do not overlap
			in, out := 2*float64(k), 2*float64(n-1-k)
			if up.x < low.x {
				x0 = float64(up.x+up.w) - 1.5 - out
				x1 = float64(low.x) + 1.5 + in
			} else {
				x0 = float64(up.x) + 1.5 + in
				x1 = float64(low.x+low.w) - 1.5 - out
			}
		}
		m.tail, m.head = [2]float64{x0, y0}, [2]float64{x1, y1}
		if m.from != up {
			m.tail, m.head = m.head, m.tail
		}
		mx, my := (x0+x1)/2, (y0+y1)/2
		if k < n/2 {
			m.said, m.anchor = [2]float64{mx - 1, my}, "end"
		} else {
			m.said, m.anchor = [2]float64{mx + 1, my}, "start"
		}
	}
}

// across lays the messages between a box and one beside it, straight
// from the side of one to the side of the other. The first half of them
// have their labels above, the rest below.
func across(a, b *box, ms []*arrow) {
	l, r := a, b
	if b.x < a.x {
		l, r = b, a
	}
	x0, x1 := float64(l.x+l.w)-0.5, float64(r.x)+0.5
	lo := float64(max(l.y, r.y)) + 1.5
	hi := float64(min(l.y+l.h, r.y+r.h)) - 1.5
	n := len(ms)
	for k, m := range ms {
		y := lo + (hi-lo)*float64(k+1)/float64(n+1)
		m.tail, m.head = [2]float64{x0, y}, [2]float64{x1, y}
		if m.from != l {
			m.tail, m.head = m.head, m.tail
		}
		m.said, m.anchor = [2]float64{(x0 + x1) / 2, y - 0.8}, "middle"
		if k >= (n+1)/2 {
			m.said[1] = y + 0.8
		}
	}
}

// span is where m's word stands, in cells: its first column, and how
// many it takes; runes write it there, and the page's text falls inside.
func (m *arrow) span() (x0, n int) {
	n = runes(m.label)
	switch m.anchor {
	case "end":
		return int(m.said[0]) - n, n
	case "start":
		return int(m.said[0]) + 1, n
	}
	return int(math.Round(m.said[0] - float64(n)/2)), n
}

// shift moves the figure d cells right: every box, and every message's
// line and word. The lanes' labels stay in the column at left.
func (f *Figure) shift(d int) {
	var move func(b *box)
	move = func(b *box) {
		b.x += d
		for _, k := range b.kids {
			move(k)
		}
	}
	for _, l := range f.lanes {
		move(l)
	}
	for _, m := range f.arrows {
		m.tail[0] += float64(d)
		m.head[0] += float64(d)
		m.said[0] += float64(d)
	}
	f.left += d
	f.w += d
}

// lay sizes every box and places it: the lanes one under another, as far
// apart as boxes stacked are, all as wide as the widest, right of the
// column their labels take.
func (f *Figure) lay() {
	w := 0
	for _, l := range f.lanes {
		l.size(0, f)
		w = max(w, l.w)
		left := []string{l.label}
		if len(l.kids) > 0 {
			left = append(left, l.lines...)
		}
		f.left = max(f.left, widest(left...)+2)
	}
	y := 0
	for i, l := range f.lanes {
		if i > 0 {
			y += f.apart(f.lanes[i-1], l, false)
		}
		l.place(f.left, y, w, l.h, 0, f)
		y += l.h
	}
	f.w, f.h = f.left+w, y
}

// route lays each message's line between its two boxes: down when one is
// above the other, across when they stand side by side. Messages between
// the same two boxes share the room between them evenly. A box cannot
// send to one it holds or is held by: that is an error at the message.
func (f *Figure) route() error {
	type pair struct{ a, b *box }
	var pairs []pair
	along := map[pair][]*arrow{}
	for _, m := range f.arrows {
		p := pair{m.from, m.to}
		if _, ok := along[pair{m.to, m.from}]; ok {
			p = pair{m.to, m.from}
		}
		if _, ok := along[p]; !ok {
			pairs = append(pairs, p)
		}
		along[p] = append(along[p], m)
	}
	for _, p := range pairs {
		a, b, ms := p.a, p.b, along[p]
		switch {
		case a.y+a.h <= b.y || b.y+b.h <= a.y:
			down(a, b, ms)
		case a.x+a.w <= b.x || b.x+b.w <= a.x:
			across(a, b, ms)
		default:
			return fmt.Errorf("%s: → %s: one box holds the other",
				ms[0].at, ms[0].to.label)
		}
	}
	return nil
}

// fit makes room for what the messages say: a word beside a line near an
// edge would run past the figure, and be cut there. Past the left edge
// the whole figure moves right; past the right, it widens.
func (f *Figure) fit() {
	left, right := 0, f.w
	for _, m := range f.arrows {
		x0, n := m.span()
		if n == 0 {
			continue
		}
		left, right = min(left, x0), max(right, x0+n+1)
	}
	if left < 0 {
		f.shift(-left)
		right -= left
	}
	f.w = max(f.w, right)
}
