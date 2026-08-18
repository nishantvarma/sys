func (m *fm) toggleSel() {
	if c := m.cur(); c != "" {
		if m.sel[c] {
			delete(m.sel, c)
		} else {
			m.sel[c] = true
		}
	}
}
