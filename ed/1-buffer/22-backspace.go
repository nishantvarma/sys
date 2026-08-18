func (b *buffer) backspace(p pos) pos {
	if p.col > 0 {
		ln := slices.Clone(b.lines[p.line]) // history holds the old
		b.lines[p.line] = slices.Delete(ln, p.col-1, p.col)
		b.dirty = true
		return pos{p.line, p.col - 1}
	}
	if p.line == 0 {
		return p
	}
	prev := b.lines[p.line-1]
	np := pos{p.line - 1, len(prev)}
	b.lines[p.line-1] = append(slices.Clone(prev), b.lines[p.line]...)
	b.lines = slices.Delete(b.lines, p.line, p.line+1)
	b.dirty = true
	return np
}
