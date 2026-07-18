func (e *editor) open(above bool) {
	e.b.snapshot()
	row := e.head().line
	if above {
		row--
	}
	e.startInsert(e.b.openLine(row))
}
