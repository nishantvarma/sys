func (m *fm) rename() {
	c := m.cur()
	if c == "" {
		return
	}
	if name, ok := m.prompt("mv: ", filepath.Base(c)); ok && name != "" {
		if m.catch(os.Rename(c, filepath.Join(m.cwd, name))) {
			m.ls() // the cursor keeps a name: land on the new one
			m.land(filepath.Base(name))
		}
	}
}
