// render is ed's screen, w by h, in tty's palette, as the samples drew a
// source: on top, where there is room, the head, the file and the label
// of the line the cursor is on, as fm's head says an item, a rule, and
// what is lit; then each line after its number, the cursor's item lit
// across (block); then the status, and the cursor put where it is, shaped
// by mode.
func (v *term) render(
	n *outline.Node,
	status string,
	state []string,
	w, h int,
) string {
	line := curLine(n)
	top := 0
	if h >= 12 { // as sys/view keeps its chrome
		top = 5
	}
	body := max(1, h-1-top) // screen rows for lines
	v.scroll(n, line, body, max(1, w-gutter))
	a, z := block(n, line)
	var b strings.Builder
	b.WriteString(tty.Home)
	if top > 0 {
		text := strings.TrimSpace(n.Kids[line].Text)
		label, _, _ := strings.Cut(text, "  ")
		tty.Line(&b, view.Head(userHome, n.ID, label, nil, w))
		tty.Line(&b, tty.Paint(strings.Repeat("─", w), tty.Rule,
			tty.Ground, false))
		tty.Line(&b, blank(w))
		tty.Line(&b, written(n.ID, a, z, w))
		tty.Line(&b, blank(w))
	}
	for r := 0; r < body; r++ {
		idx := v.top + r
		if idx < len(n.Kids) {
			lit, on := a <= idx && idx <= z, idx == line
			tty.Line(&b, v.paintLine(n.Kids[idx], idx, lit, on, w))
		} else {
			tty.Line(&b, blank(w))
		}
	}
	cur := n.Kids[line]
	col := cursorCol(cur)
	b.WriteString(statusLine(n, state, status, line, col, w, h))
	// Place the real cursor at the head's visual column, shaped by mode.
	// scroll has already put that column inside the window.
	row := top + line - v.top + 1
	x := gutter + visualCol([]rune(cur.Text), col) - v.off + 1
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
