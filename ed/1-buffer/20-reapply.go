// reapply reverses the last restore — the redo half of undo. It moves the
// current lines onto the undo stack and pops the last undone state back.
func (b *buffer) reapply() bool {
	if len(b.redo) == 0 {
		return false
	}
	b.undo = append(b.undo, b.lines)
	b.lines = b.redo[len(b.redo)-1]
	b.redo = b.redo[:len(b.redo)-1]
	b.dirty = true
	return true
}
