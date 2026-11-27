// render is ed's screen, w by h, in tty's palette: each line after its
// number, the runes under the selection on a ground of their own, and the
// status at the bottom; then where the cursor goes, shaped by mode. A
// row a string, for tty's Frame to write what changed.
func (v *term) render(
	n *outline.Node,
	status string,
	state []string,
	w, h int,
) ([]string, string) {
	line := curLine(n)
	g := gutter(len(n.Kids))
	body := max(1, h-1) // screen rows for lines
	v.scroll(n, line, body, max(1, w-g))
	rows := make([]string, 0, body+1)
	for r := 0; r < body; r++ {
		idx := v.top + r
		if idx >= len(n.Kids) {
			rows = append(rows, blank(w))
			continue
		}
		rows = append(rows, v.paintLine(n.Kids[idx], idx, g, w))
	}
	cur := n.Kids[line]
	col := cursorCol(cur)
	rows = append(rows, statusLine(n, state, status, line, col, w))
	// scroll has already put the cursor's column inside the window
	row := line - v.top + 1
	x := g + visualCol([]rune(cur.Text), col) - v.off + 1
	return rows, tty.Goto(row, x) + tty.Cnorm + cursorShape(state[0])
}
