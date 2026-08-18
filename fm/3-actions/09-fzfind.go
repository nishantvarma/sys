func (m *fm) fzfind() {
	if p := m.pick("f"); p != "" {
		m.reveal(p, 0)
	}
}
