func splitRunes(rs []rune) [][]rune {
	out := [][]rune{{}}
	for _, r := range rs {
		if r == '\n' {
			out = append(out, []rune{})
		} else {
			out[len(out)-1] = append(out[len(out)-1], r)
		}
	}
	return out
}
