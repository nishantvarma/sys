func (m *fm) create(label string, mk func(string) error) {
	if name, ok := m.prompt(label+": ", ""); ok && name != "" {
		if m.catch(mk(filepath.Join(m.cwd, name))) {
			m.ls() // land on it: cur follows what a verb made
			m.land(filepath.Base(name))
		}
	}
}
