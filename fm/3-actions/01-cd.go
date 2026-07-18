func (m *fm) cd(p string) {
	p, _ = filepath.Abs(p)
	// commit navigation state only on success
	if err := os.Chdir(p); err != nil {
		m.flash(err.Error())
		return
	}
	m.pos[m.cwd] = m.idx
	m.last, m.cwd = m.cwd, p
	m.idx = m.pos[m.cwd]
	m.title("fm:" + m.cwd)
	m.sel = map[string]bool{}
	zoxideAdd(m.cwd)
}
