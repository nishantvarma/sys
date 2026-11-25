// say writes the row to cur, the noun every verb reads when it is given
// none. Only while this pane holds the eye: a background fm — scrolled,
// or woken by a resize — must not take the shell's this. state.Write puts
// it whole, so a reader gets the old address or the new, never an empty
// one. The log is c's alone: a view moves this on every j, and at a line a
// keypress back would be unusable.
func (m *fm) say() {
	c := m.cur()
	if !m.on || c == "" || c == m.said {
		return
	}
	if state.Write(m.this, c) == nil {
		m.said = c
	}
}
