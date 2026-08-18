func (e *editor) insertAt(p pos) {
	e.b.snapshot(e.caret)
	e.startInsert(p)
}
