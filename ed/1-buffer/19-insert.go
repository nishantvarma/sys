func (b *buffer) insert(p pos, r rune) pos {
	ln := slices.Clone(b.lines[p.line]) // history holds the old one
	b.lines[p.line] = slices.Insert(ln, p.col, r)
	b.dirty = true
	return pos{p.line, p.col + 1}
}
