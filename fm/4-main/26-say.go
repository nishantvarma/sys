// say writes the row to cur, the noun every verb reads when it is given
// none. Only while this pane holds the eye: a background fm — scrolled,
// or woken by a resize — must not take the shell's this. state.Write puts
// it whole, so a reader gets the old address or the new, never an empty
// one. The log is c's alone: a view moves this on every j, and at a line a
// keypress back would be unusable. cur that names the row already, by any
// path, is left as it is, for it may carry a line, a save's or a click's.
func (m *fm) say() {
	c := m.cur()
	if !m.on || c == "" || c == m.said {
		return
	}
	if p, _ := outline.Address(state.Read(m.this)); same(p, c) {
		m.said = c
		return
	}
	if state.Write(m.this, c) == nil {
		m.said, m.heard = c, c // its own word, no news to follow
	}
}
