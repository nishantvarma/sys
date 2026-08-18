func (m *fm) fzsearch() {
	pat, ok := m.prompt("rg: ", "")
	if !ok || pat == "" {
		return
	}
	m.revealAddr(m.pick("s", "-n", pat))
}
