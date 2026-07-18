func (m *fm) create(label string, mk func(string) error) {
	if name, ok := m.prompt(label+": ", nil, ""); ok && name != "" {
		m.catch(mk(filepath.Join(m.cwd, name)))
	}
}
