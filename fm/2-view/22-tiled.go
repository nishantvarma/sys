// tiled is a screen of rows as tiles, a grid read left to right: the
// header, as many lines of tiles as fit, the cursor's line kept in the
// middle when it can be, a count of the tiles that did not fit, and the
// status, as render is for a list.
func tiled(head *outline.Node, rows []row, status string, w, h int) string {
	cols, cw := grid(rows, w)
	var lines []string
	for i := 0; i < len(rows); i += cols {
		var s strings.Builder
		for j := i; j < min(i+cols, len(rows)); j++ {
			s.WriteString(tile(j, rows[j], cw))
		}
		lines = append(lines, s.String())
	}
	var b strings.Builder
	b.WriteString(tty.Home)
	tty.Line(&b, tty.Clip(header(head, w), w))
	n := h - 2
	if len(lines) > n {
		n--
	}
	n = max(0, n)
	off := max(0, min(curRow(rows)/cols-n/2, len(lines)-n))
	vis := lines[off:min(off+n, len(lines))]
	for _, l := range vis {
		tty.Line(&b, tty.Clip(l, w))
	}
	extra := n - len(vis)
	if len(lines) > n {
		seen := min(len(rows), (off+len(vis))*cols) - off*cols
		tty.Line(&b, tty.Dim(fmt.Sprintf(" +%d", len(rows)-seen)))
		extra--
	}
	for i := 0; i < extra; i++ {
		tty.Line(&b, "")
	}
	b.WriteString(tty.Status(h, tty.Clip(status, w), false))
	return b.String()
}
