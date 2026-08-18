func (m *fm) mv(d int) {
	if len(m.files) > 0 {
		m.idx = max(0, min(m.idx+d, len(m.files)-1))
	}
}
