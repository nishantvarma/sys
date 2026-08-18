// reapply reverses the last restore — the redo half of undo, and its mirror:
// the current state goes onto the undo stack and the last undone one pops
// back, caret and all.
func (b *buffer) reapply(c sel) (sel, bool) {
	if len(b.redo) == 0 {
		return c, false
	}
	s := b.redo[len(b.redo)-1]
	b.undo = append(b.undo, snap{b.lines, c})
	b.redo = b.redo[:len(b.redo)-1]
	b.lines = s.lines
	b.dirty = true
	return s.caret, true
}
