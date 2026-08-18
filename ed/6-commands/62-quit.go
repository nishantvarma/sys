func (e *editor) quit() {
	if e.b.dirty && !e.armed {
		e.armed = true
		e.msg = "modified — q again to discard"
		return
	}
	e.done = true
}
