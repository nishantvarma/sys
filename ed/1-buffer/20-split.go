// split breaks the line at p, returning the start of the new line.
func (b *buffer) split(p pos) pos {
	ln := b.lines[p.line]
	head := slices.Clone(ln[:p.col])
	tail := slices.Clone(ln[p.col:])
	b.lines[p.line] = head
	b.lines = slices.Insert(b.lines, p.line+1, tail)
	b.dirty = true
	return pos{p.line + 1, 0}
}
