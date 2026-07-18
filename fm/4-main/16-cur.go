func (m *fm) cur() string {
	if len(m.files) == 0 {
		return ""
	}
	return m.files[m.idx].path
}
