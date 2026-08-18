// wordBack lands on the start of the current or previous word.
func wordBack(b *buffer, p pos) pos {
	n := back(b, p)
	if n == p {
		return p
	}
	p = n
	for classAt(b, p) == clSpace {
		n := back(b, p)
		if n == p {
			return p
		}
		p = n
	}
	cls := classAt(b, p)
	for {
		n := back(b, p)
		if n == p || classAt(b, n) != cls {
			return p
		}
		p = n
	}
}
