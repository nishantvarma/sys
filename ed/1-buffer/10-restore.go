func (b *buffer) restore() bool {
	if len(b.undo) == 0 {
		return false
	}
	b.redo = append(b.redo, b.lines)
	b.lines = b.undo[len(b.undo)-1]
	b.undo = b.undo[:len(b.undo)-1]
	b.dirty = true
	return true
}
