func (m *fm) ls() {
	m.stale.Store(false)
	m.reads++
	if m.watcher != nil {
		m.watcher.Follow(m.cwd)
	}
	name := m.curName()
	m.files = entries(m.cwd, m.hidden)
	// the cursor keeps its name across the rebuild, else its row
	if i := m.at(name); i >= 0 {
		m.idx = i
	} else {
		m.idx = max(0, min(m.idx, len(m.files)-1))
	}
}
