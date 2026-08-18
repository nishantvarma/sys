// up moves up a line, aiming at goal rather than the current column,
// and restores goal after — place would otherwise narrow it to
// whatever a short line clamped it to.
func (e *editor) up() {
	g := e.goal
	e.moveHead(-1, g-e.head().col)
	e.goal = g
}
