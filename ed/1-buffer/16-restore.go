// restore steps back one state and hands its caret back. c is the live
// caret, which goes onto the redo stack: it is where the undone edit left
// the cursor, and where redo should put it again.
func (b *buffer) restore(c sel) (sel, bool) {
	if len(b.undo) == 0 {
		return c, false
	}
	s := b.undo[len(b.undo)-1]
	b.redo = append(b.redo, snap{b.lines, c})
	b.undo = b.undo[:len(b.undo)-1]
	b.lines = s.lines
	b.dirty = true
	return s.caret, true
}
