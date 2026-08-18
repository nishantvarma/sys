type buffer struct {
	lines [][]rune
	path  string
	dirty bool
	undo  []snap
	redo  []snap
}
