type buffer struct {
	lines [][]rune
	path  string
	dirty bool
	undo  [][][]rune
	redo  [][][]rune
}
