func (m *fm) edit() {
	c := m.cur()
	if c == "" || !isFile(c) {
		return
	}
	if n := m.hitLine(c); n > 0 {
		m.spawn(true, false, cmdEdit, "+"+strconv.Itoa(n))
		return
	}
	m.spawn(true, false, cmdEdit)
}
