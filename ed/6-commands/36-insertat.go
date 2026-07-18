func (e *editor) insertAt(p pos) {
	e.b.snapshot()
	e.startInsert(p)
}
