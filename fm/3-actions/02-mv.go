func (m *fm) mv(d int) {
	if len(m.files) > 0 {
		m.idx = tty.Clamp(m.idx+d, 0, len(m.files)-1)
	}
}
