func (e *editor) setCol(c int) {
	e.point(pos{e.head().line, c})
}
