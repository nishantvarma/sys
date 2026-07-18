func (m *fm) goto_() {
	s, ok := m.prompt("goto: ", m.complete, "")
	if !ok || s == "" {
		return
	}
	p, _ := filepath.Abs(expand(s))
	if r, err := filepath.EvalSymlinks(p); err == nil {
		p = r
	}
	switch {
	case isDir(p):
		m.cd(p)
	case isFile(p):
		m.spawn(false, false, cmdOpen, p)
	}
}
