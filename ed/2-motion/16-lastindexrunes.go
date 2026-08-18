func lastIndexRunes(ln, pat []rune, lo, hi int) int {
	for i := hi - len(pat); i >= lo; i-- {
		if slices.Equal(ln[i:i+len(pat)], pat) {
			return i
		}
	}
	return -1
}
