// down mirrors up.
func (e *editor) down() {
	g := e.goal
	e.moveHead(1, g-e.head().col)
	e.goal = g
}
