func fwd(b *buffer, p pos) pos {
	if p.col < len(b.lines[p.line]) {
		return pos{p.line, p.col + 1}
	}
	if p.line < b.last() {
		return pos{p.line + 1, 0}
	}
	return p
}
