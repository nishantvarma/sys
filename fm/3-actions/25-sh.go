func (m *fm) sh() {
	if execable(filepath.Join(m.ctx(), "sh")) {
		m.detach("./sh")
	} else {
		m.detach(shellCmd...)
	}
}
