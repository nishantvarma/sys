func paraBack(b *buffer, p pos) pos {
	l := p.line
	for l > 0 && len(b.lines[l]) == 0 {
		l-- // skip the blank run the cursor is in
	}
	for l > 0 && len(b.lines[l]) != 0 {
		l-- // scan to the previous blank line
	}
	return pos{l, 0}
}
