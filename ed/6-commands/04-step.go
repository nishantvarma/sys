// step walks the cursor along a motion. A landing on a line break is a spot
// normal mode can't rest on, and clamp would resolve it backwards — undoing
// a forward wrap — so step repeats the motion instead, in its own direction.
func (e *editor) step(m func(*buffer, pos) pos) {
	p := m(e.b, e.head())
	if !e.mode.pastEnd() && p.col > 0 && p.col == len(e.b.lines[p.line]) {
		p = m(e.b, p)
	}
	e.point(p)
}
