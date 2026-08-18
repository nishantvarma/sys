// text returns the runes in the inclusive range [a, c], with '\n' between
// lines. An end column at line length includes the line break, as runeAt
// reads it.
func (b *buffer) text(a, c pos) []rune {
	a, c = order(a, c)
	if a.line == c.line {
		ln := b.lines[a.line]
		out := append([]rune(nil), ln[a.col:end(ln, c.col)]...)
		if c.col >= len(ln) {
			out = append(out, '\n')
		}
		return out
	}
	var out []rune
	first := b.lines[a.line]
	out = append(out, first[min(a.col, len(first)):]...)
	out = append(out, '\n')
	for l := a.line + 1; l < c.line; l++ {
		out = append(out, b.lines[l]...)
		out = append(out, '\n')
	}
	last := b.lines[c.line]
	out = append(out, last[:end(last, c.col)]...)
	if c.col >= len(last) {
		out = append(out, '\n')
	}
	return out
}
