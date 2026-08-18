// paste inserts text (which may contain '\n') at p, returning the end
// position.
func (b *buffer) paste(p pos, text []rune) pos {
	parts := splitRunes(text)
	ln := b.lines[p.line]
	pre := slices.Clone(ln[:p.col])
	post := slices.Clone(ln[p.col:])
	if len(parts) == 1 {
		b.lines[p.line] = append(append(pre, parts[0]...), post...)
		b.dirty = true
		return pos{p.line, p.col + len(parts[0])}
	}
	rows := [][]rune{append(pre, parts[0]...)}
	for _, m := range parts[1 : len(parts)-1] {
		rows = append(rows, slices.Clone(m))
	}
	tail := parts[len(parts)-1]
	rows = append(rows, append(slices.Clone(tail), post...))
	b.lines = slices.Replace(b.lines, p.line, p.line+1, rows...)
	b.dirty = true
	return pos{p.line + len(parts) - 1, len(tail)}
}
