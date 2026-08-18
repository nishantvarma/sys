// reset takes over src's lines. reload reads a fresh buffer and hands it
// here rather than reaching into it — the spine belongs to this file and
// its neighbours, not to a command. What was read is what is on disk, so
// the buffer comes back clean and save may write over it; history keeps
// the lines it replaced, with the caret c that was in them. A read is no
// edit, so it takes no snapshot and edits stays put: . repeats the last
// edit, never a read.
func (b *buffer) reset(src *buffer, c sel) {
	b.undo = append(b.undo, snap{b.lines, c})
	b.redo = nil
	b.lines = src.lines
	b.disk = src.disk
	b.dirty = false
}
