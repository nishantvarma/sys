func back(b *buffer, p pos) pos {
	if p.col > 0 {
		return pos{p.line, p.col - 1}
	}
	if p.line > 0 {
		return pos{p.line - 1, len(b.lines[p.line-1])}
	}
	return p
}
