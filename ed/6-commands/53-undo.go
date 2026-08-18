func (e *editor) undo() {
	if c, ok := e.b.restore(e.caret); ok {
		e.recall(c)
	} else {
		e.msg = "nothing to undo"
	}
}
