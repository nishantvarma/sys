// grid is how many tiles fit across w, and how wide each tile's label is:
// the widest, up to 24, so that the tiles line up in columns. A tile is a
// space, its number, a space, its label and a space.
func grid(rows []row, w int) (cols, cw int) {
	for _, r := range rows {
		cw = max(cw, tty.Width(label(r.n)))
	}
	cw = min(cw, 24)
	return max(1, w/(cw+5)), cw
}
