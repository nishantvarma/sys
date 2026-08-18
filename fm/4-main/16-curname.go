func (m *fm) curName() string {
	if len(m.files) == 0 {
		return ""
	}
	return m.files[m.idx].name
}
