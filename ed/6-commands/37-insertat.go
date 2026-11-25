func (e *editor) insertAt(p pos) {
	e.startInsert(p)
	e.owed = true
}
