// indent is how far a line stands in, in visual columns; a blank line
// stands in not at all.
func indent(s string) int {
	r := []rune(s)
	i := 0
	for i < len(r) && (r[i] == ' ' || r[i] == '\t') {
		i++
	}
	if i == len(r) {
		return 0
	}
	return visualCol(r, i)
}
