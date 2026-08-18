// end clamps an inclusive column c to an exclusive slice bound within ln.
func end(ln []rune, c int) int {
	if c+1 > len(ln) {
		return len(ln)
	}
	return c + 1
}
