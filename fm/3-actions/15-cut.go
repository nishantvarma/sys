func (m *fm) cut() {
	m.clip, m.cutting, m.sel = m.targets(), true, map[string]bool{}
}
