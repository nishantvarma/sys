func (m *fm) rename() {
	c := m.cur()
	if c == "" {
		return
	}
	if name, ok := m.prompt("mv: ", nil, filepath.Base(c)); ok &&
		name != "" {
		m.catch(os.Rename(c, filepath.Join(m.cwd, name)))
	}
}
