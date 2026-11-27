package picture

// inner is the innermost of b and the boxes in it that holds the cell at
// x, y; nil when b does not hold it.
func (b *box) inner(x, y int) *box {
	if x < b.x || x >= b.x+b.w || y < b.y || y >= b.y+b.h {
		return nil
	}
	for _, k := range b.kids {
		if in := k.inner(x, y); in != nil {
			return in
		}
	}
	return b
}

// covers is whether m's word stands on the cell at x, y.
func (m *arrow) covers(x, y int) bool {
	x0, n := m.span()
	return n > 0 && y == int(m.said[1]) && x >= x0 && x < x0+n
}

// At is the address a cell of Runes names, column x and row y from 0: the
// message whose word stands there, for runes and the page draw words over
// the rest, else the innermost box that holds it; "" for none. A click on
// a picture names what it falls on.
func (f *Picture) At(x, y int) string {
	for _, m := range f.arrows {
		if m.covers(x, y) {
			return m.at
		}
	}
	for _, l := range f.lanes {
		if b := l.inner(x, y); b != nil {
			return b.at
		}
	}
	return ""
}
