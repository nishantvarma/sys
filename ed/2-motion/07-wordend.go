// wordEnd lands on the end of the current or next word (or punctuation run).
func wordEnd(b *buffer, p pos) pos {
	n := fwd(b, p)
	if n == p {
		return p
	}
	p = n
	for classAt(b, p) == clSpace {
		n := fwd(b, p)
		if n == p {
			return p
		}
		p = n
	}
	cls := classAt(b, p)
	for {
		n := fwd(b, p)
		if n == p || classAt(b, n) != cls {
			return p
		}
		p = n
	}
}
