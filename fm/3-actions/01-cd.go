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
	if !under(m.cwd, m.root) { // out past the top: the top comes along
		if under(m.root, m.cwd) { // the old top stays open below
			m.open[m.root] = true
		}
		m.root = m.cwd
	}
	m.ls()
	m.land(m.pos[m.cwd])
	m.title("fm:" + m.cwd)
	m.sel = map[string]bool{}
	m.item = "" // a jump leaves any item
	zoxideAdd(m.cwd)
}
