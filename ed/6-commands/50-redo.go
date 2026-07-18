func (e *editor) redo() {
	if e.b.reapply() {
		e.place(e.head())
	} else {
		e.msg = "nothing to redo"
	}
}
