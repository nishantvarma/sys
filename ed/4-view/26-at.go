// at is the line a cell shows, by its index, and the rune there counted
// from 1, its end past the line's end; x and y count from 1, as the
// terminal counts them. -1 for the status line, or before a draw.
func (v *term) at(x, y int) (int, int) {
	n := v.last
	_, h := v.t.Size()
	if n == nil || len(n.Kids) == 0 || y < 1 || y >= h {
		return -1, 0
	}
	i := min(v.top+y-1, len(n.Kids)-1)
	line := []rune(n.Kids[i].Text)
	col := x - 1 - gutter(len(n.Kids)) + v.off // the visual column
	r := 0
	for r < len(line) && visualCol(line, r+1) <= col {
		r++
	}
	return i, r + 1
}
