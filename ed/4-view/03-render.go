func render(f frame) string {
	var b strings.Builder
	b.WriteString(tty.Home)
	rows := f.h - 1
	for r := 0; r < rows; r++ {
		idx := f.top + r
		if idx < len(f.lines) {
			tty.Line(&b, paintLine(f, idx))
		} else {
			tty.Line(&b, "")
		}
	}
	b.WriteString(statusLine(f))
	// Place the real cursor at the head's visual column, shaped by mode.
	// scroll has already put that column inside the window.
	row := f.head.line - f.top + 1
	col := visualCol(f.lines[f.head.line], f.head.col) - f.off + 1
	b.WriteString(
		fmt.Sprintf(
			"\x1b[%d;%dH%s%s",
			row,
			col,
			tty.Cnorm,
			cursorShape(f.mode),
		),
	)
	return b.String()
}
