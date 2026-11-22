// cursorCol is the rune column of the cursor in line ln: the runes before
// the run that holds it.
func cursorCol(ln *outline.Node) int {
	col := 0
	for _, r := range ln.Kids {
		if r.Cur {
			break
		}
		if r.Text != "\n" {
			col += len([]rune(r.Text))
		}
	}
	return col
}
