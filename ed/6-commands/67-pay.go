// pay takes the snapshot an insert owes before its first edit: one begun
// by i, a, A or a drag, or one an undo or a redo went back in, so that an
// insert that changes nothing leaves no undo behind.
func (e *editor) pay() {
	if e.owed {
		e.b.snapshot(e.caret)
		e.owed = false
	}
}
