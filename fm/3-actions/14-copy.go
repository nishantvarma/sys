func (m *fm) copy() {
	m.clip, m.cutting, m.sel = m.targets(), false, map[string]bool{}
}
