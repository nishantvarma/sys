// page moves the cursor a screenful in direction d (+1 down, -1 up),
// keeping a line of overlap, and preserves goal like up/down.
func (e *editor) page(d int) {
	_, h := e.t.Size()
	g := e.goal
	e.moveHead(d*(h-2), g-e.head().col)
	e.goal = g
}
