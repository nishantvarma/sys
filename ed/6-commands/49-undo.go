func (e *editor) undo() {
	if e.b.restore() {
		e.place(e.head())
	} else {
		e.msg = "nothing to undo"
	}
}
