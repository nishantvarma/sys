func (b *buffer) insert(p pos, r rune) pos {
	b.lines[p.line] = slices.Insert(b.lines[p.line], p.col, r)
	b.dirty = true
	return pos{p.line, p.col + 1}
}
