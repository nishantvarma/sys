// wordFwd lands on the start of the next word (or punctuation run).
func wordFwd(b *buffer, p pos) pos {
	cls := classAt(b, p)
	for {
		n := fwd(b, p)
		if n == p {
			return p
		}
		p = n
		if classAt(b, p) != cls {
			break
		}
	}
	for classAt(b, p) == clSpace {
		n := fwd(b, p)
		if n == p {
			break
		}
		p = n
	}
	return p
}
