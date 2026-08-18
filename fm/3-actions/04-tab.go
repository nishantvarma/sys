func (m *fm) tab() {
	switch {
	case m.alt != "":
		prev := m.cwd
		m.cd(m.alt)
		m.alt = prev
	case m.last != "":
		m.cd(m.last)
	}
}
