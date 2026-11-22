func (v *term) render(
	n *outline.Node,
	status string,
	state []string,
	w, h int,
) string {
	line := curLine(n)
	v.scroll(n, line, h-1, w)
	var b strings.Builder
	b.WriteString(tty.Home)
	body := h - 1 // screen rows above the status line
	for r := 0; r < body; r++ {
		idx := v.top + r
		if idx < len(n.Kids) {
			tty.Line(&b, v.paintLine(n.Kids[idx], w))
		} else {
			tty.Line(&b, "")
		}
	}
	cur := n.Kids[line]
	col := cursorCol(cur)
	b.WriteString(statusLine(n, state, status, line, col, w, h))
	// Place the real cursor at the head's visual column, shaped by mode.
	// scroll has already put that column inside the window.
	row := line - v.top + 1
	x := visualCol([]rune(cur.Text), col) - v.off + 1
	b.WriteString(
		fmt.Sprintf(
			"\x1b[%d;%dH%s%s",
			row,
			x,
			tty.Cnorm,
			cursorShape(state[0]),
		),
	)
	return b.String()
}
