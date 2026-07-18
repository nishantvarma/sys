func (b *buffer) snapshot() {
	cp := make([][]rune, len(b.lines))
	for i, ln := range b.lines {
		cp[i] = slices.Clone(ln)
	}
	b.undo = append(b.undo, cp)
	if len(b.undo) > 256 {
		b.undo = b.undo[1:]
	}
	b.redo = nil // a fresh edit branches history; the redo stack is stale
}
