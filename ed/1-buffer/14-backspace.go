func (b *buffer) backspace(p pos) pos {
	if p.col > 0 {
		ln := slices.Delete(b.lines[p.line], p.col-1, p.col)
		b.lines[p.line] = ln
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
