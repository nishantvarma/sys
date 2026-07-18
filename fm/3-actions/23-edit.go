func (m *fm) edit() {
	if c := m.cur(); c != "" && isFile(c) {
		m.spawn(true, false, cmdEdit)
	}
}
