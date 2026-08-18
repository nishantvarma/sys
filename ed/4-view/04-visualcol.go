// visualCol is the screen column of rune index col, counting tab stops.
func visualCol(line []rune, col int) int {
	v := 0
	for i := 0; i < col && i < len(line); i++ {
		if line[i] == '\t' {
			v += tabWidth - v%tabWidth
		} else {
			v++
		}
	}
	return v
}
