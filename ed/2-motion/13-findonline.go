// findOnLine scans line p.line from just past p for r, forward or backward.
// ok is false if r doesn't occur in that direction on the line.
func findOnLine(b *buffer, p pos, r rune, forward bool) (pos, bool) {
	ln := b.lines[p.line]
	if forward {
		for c := p.col + 1; c < len(ln); c++ {
			if ln[c] == r {
				return pos{p.line, c}, true
			}
		}
		return p, false
	}
	for c := p.col - 1; c >= 0; c-- {
		if ln[c] == r {
			return pos{p.line, c}, true
		}
	}
	return p, false
}
