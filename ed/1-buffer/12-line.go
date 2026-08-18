// line hands out row i. The result is read-only: it is the very slice
// history holds, so writing through it would edit states already undone
// past. Callers that mean to change a line go through mapLine.
func (b *buffer) line(i int) []rune { return b.lines[i] }
