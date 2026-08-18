// indexRunes/lastIndexRunes find pat within ln[lo:hi], preferring the match
// nearest lo or nearest hi respectively. Both return -1 on no match.
func indexRunes(ln, pat []rune, lo, hi int) int {
	for i := lo; i+len(pat) <= hi; i++ {
		if slices.Equal(ln[i:i+len(pat)], pat) {
			return i
		}
	}
	return -1
}
