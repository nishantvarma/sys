// wordAt is the word around rune i of line, a run of runes between blanks,
// as tmux takes the word under a click (`tmux/conf:/word-separators/`);
// empty on a blank.
func wordAt(line []rune, i int) string {
	space := func(j int) bool { return line[j] == ' ' || line[j] == '\t' }
	if i < 0 || i >= len(line) || space(i) {
		return ""
	}
	a, z := i, i
	for a > 0 && !space(a-1) {
		a--
	}
	for z+1 < len(line) && !space(z+1) {
		z++
	}
	return string(line[a : z+1])
}
