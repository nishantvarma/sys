// flash shows msg in the next frame's status line; it clears on the next
// keypress, like ed's — no blocking sleep.
func (m *fm) flash(msg string) { m.msg = msg }
