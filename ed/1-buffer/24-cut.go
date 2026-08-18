// cut deletes the inclusive range [a, c] and returns what it removed. An end
// column at line length includes the line break (as runeAt reads it), so the
// next line joins on — or, at the buffer's edge, the emptied line is
// dropped.
func (b *buffer) cut(a, c pos) []rune {
	a, c = order(a, c)
	gone := b.text(a, c)
	first := b.lines[a.line]
	pre := append([]rune(nil), first[:min(a.col, len(first))]...)
	last := b.lines[c.line]
	out := make([][]rune, 0, len(b.lines))
	out = append(out, b.lines[:a.line]...)
	switch {
	case c.col < len(last): // ends on a rune: keep the line's tail
		out = append(out, append(pre, last[c.col+1:]...))
		out = append(out, b.lines[c.line+1:]...)
	case c.line < b.last(): // line break included: join the next line on
		out = append(out, append(pre, b.lines[c.line+1]...))
		out = append(out, b.lines[c.line+2:]...)
	case len(pre) > 0 || a.line == 0:
		// break at buffer edge: nothing to join
		out = append(out, pre)
	} // else whole trailing lines gone: drop them, dd-on-last-line style
	if len(out) == 0 {
		out = [][]rune{{}}
	}
	b.lines = out
	b.dirty = true
	return gone
}
