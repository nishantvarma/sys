func (m *fm) cd(p string) {
	p, _ = filepath.Abs(p)
	if p == m.cwd { // enter on . stays put, selection intact
		return
	}
	// commit navigation state only on success
	if err := os.Chdir(p); err != nil {
		m.flash(err.Error())
		return
	}
	m.pos[m.cwd] = m.curName()
	m.last, m.cwd = m.cwd, p
	m.ls()
	m.land(m.pos[m.cwd])
	m.title("fm:" + m.cwd)
	m.sel = map[string]bool{}
	zoxideAdd(m.cwd)
}
