func (m *fm) enter() {
	c := m.cur()
	if c == "" {
		return
	}
	if isDir(c) {
		m.cd(c)
	} else {
		m.spawn(true, false, cmdOpen)
	}
}
