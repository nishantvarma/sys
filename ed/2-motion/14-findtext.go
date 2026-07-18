// findText scans forward or backward from p for pat (a literal substring),
// wrapping around the buffer. ok is false if pat is empty or not found.
func findText(b *buffer, p pos, pat string, forward bool) (a, c pos, ok bool) {
	patR := []rune(pat)
	if len(patR) == 0 {
		return p, p, false
	}
	n := len(b.lines)
	for i := 0; i <= n; i++ {
		l := p.line + i
		if !forward {
			l = p.line - i
		}
		l = ((l % n) + n) % n
		ln := b.lines[l]
		lo, hi := 0, len(ln)
		if i == 0 {
			if forward {
				lo = p.col + 1
			} else {
				hi = p.col
			}
		}
		col := -1
		if forward {
			col = indexRunes(ln, patR, lo, hi)
		} else {
			col = lastIndexRunes(ln, patR, lo, hi)
		}
		if col >= 0 {
			return pos{l, col}, pos{l, col + len(patR) - 1}, true
		}
	}
	return p, p, false
}
