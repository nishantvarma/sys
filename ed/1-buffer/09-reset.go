// reset takes over src's lines. reload reads a fresh buffer and hands it
// here rather than reaching into it — the spine belongs to this file and
// its neighbours, not to a command. What was read is what is on disk, so
// the buffer comes back clean; history keeps the lines it replaced.
func (b *buffer) reset(src *buffer) {
	b.lines = src.lines
	b.dirty = false
}
