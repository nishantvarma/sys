// clamp keeps p inside the buffer. past allows col == line length (the
// insertion point after the last rune); normal mode passes false.
func (b *buffer) clamp(p pos, past bool) pos {
	if p.line < 0 {
		p.line = 0
	}
	if p.line > b.last() {
		p.line = b.last()
	}
	n := len(b.lines[p.line])
	if !past && n > 0 {
		n--
	}
	if p.col < 0 {
		p.col = 0
	}
	if p.col > n {
		p.col = n
	}
	return p
}
