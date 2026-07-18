func (m *fm) workflow() {
	if exe := filepath.Join(m.ctx(), cmdExec); execable(exe) {
		m.spawn(false, false, cmdOpen, exe)
	}
}
