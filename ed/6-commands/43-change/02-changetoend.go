func (e *editor) changeToEnd() {
	if e.selToEnd() {
		e.change()
		return
	}
	e.b.snapshot(e.caret)
	e.startInsert(e.head())
}
