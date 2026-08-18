func (e *editor) redo() {
	if c, ok := e.b.reapply(e.caret); ok {
		e.recall(c)
	} else {
		e.msg = "nothing to redo"
	}
}
