func (e *editor) insertAfter() {
	e.insertAt(pos{e.head().line, e.head().col + 1})
}
