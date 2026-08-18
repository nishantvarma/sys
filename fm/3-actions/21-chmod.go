func (m *fm) chmod() {
	c := m.cur()
	if c == "" {
		return
	}
	fi, err := os.Stat(c)
	if err != nil {
		return
	}
	mode := fi.Mode().Perm()
	if mode&0o111 != 0 {
		m.catch(os.Chmod(c, mode&^0o111))
	} else {
		m.catch(os.Chmod(c, mode|0o111))
	}
}
