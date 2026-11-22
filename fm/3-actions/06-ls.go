func (m *fm) ls() {
	name := m.curName()
	m.files = entries(m.cwd, m.hidden, true)
	// the cursor keeps its name across the rebuild, else its row
	if i := m.at(name); i >= 0 {
		m.idx = i
	} else {
		m.idx = max(0, min(m.idx, len(m.files)-1))
	}
}
