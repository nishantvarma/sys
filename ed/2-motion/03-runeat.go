// runeAt reads the rune at p; a column at end-of-line reads as '\n'.
func runeAt(b *buffer, p pos) (rune, bool) {
	if p.line < 0 || p.line > b.last() {
		return 0, false
	}
	ln := b.lines[p.line]
	if p.col < len(ln) {
		return ln[p.col], true
	}
	return '\n', true
}
