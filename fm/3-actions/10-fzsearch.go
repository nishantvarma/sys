func (m *fm) fzsearch() {
	if pat, ok := m.prompt("rg: ", nil, ""); ok && pat != "" {
		m.spawn(false, false, cmdFzS, pat)
	}
}
