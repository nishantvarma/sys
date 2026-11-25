// undo and redo step through history. In insert they leave the insert
// owing a snapshot, for its own is spent, and no change . can type again.
func (e *editor) undo() {
	if c, ok := e.b.restore(e.caret); ok {
		e.recall(c)
	} else {
		e.msg = "nothing to undo"
	}
	if e.mode == e.ins {
		e.owed, e.unkept = true, true
	}
}
