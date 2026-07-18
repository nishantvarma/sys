func (m *fm) search() {
	pat, ok := m.prompt("/", nil, "")
	if !ok || pat == "" {
		return
	}
	m.pat = strings.ToLower(pat)
	m.find(1)
}
