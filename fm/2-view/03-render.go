func render(head *outline.Node, rows []row, status string, w, h int) string {
	var b strings.Builder
	b.WriteString(tty.Home)
	tty.Line(&b, tty.Clip(header(head, w), w))
	n := h - 2
	if len(rows) > n {
		n--
	}
	if n < 0 {
		n = 0
	}
	off := max(0, min(curRow(rows)-n/2, len(rows)-n)) // cursor row centred
	vis := rows[off:min(off+n, len(rows))]
	for i, r := range vis {
		tty.Line(&b, tty.Clip(paint(off+i, r), w))
	}
	extra := n - len(vis)
	if len(rows) > n {
		tty.Line(&b, tty.Dim(fmt.Sprintf(" +%d", len(rows)-n)))
		extra--
	}
	for i := 0; i < extra; i++ {
		tty.Line(&b, "")
	}
	b.WriteString(tty.Status(h, tty.Clip(status, w), false))
	return b.String()
}
