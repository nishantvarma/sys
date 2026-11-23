// labelOf is where a line's label stands, as a note's item has one: from
// its first rune not blank to the first run of two blanks after it, or to
// its end; the rune indices, the second one past the label.
func labelOf(line []rune) (int, int) {
	a := 0
	for a < len(line) && (line[a] == ' ' || line[a] == '\t') {
		a++
	}
	for b := a; b+1 < len(line); b++ {
		if line[b] == ' ' && line[b+1] == ' ' {
			return a, b
		}
	}
	return a, len(line)
}
