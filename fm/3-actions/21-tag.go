func (m *fm) tag() {
	def := filepath.Base(m.cwd)
	name, ok := m.prompt(fmt.Sprintf("tag [%s]: ", def), nil, "")
	if !ok {
		return
	}
	if name == "" {
		name = def
	}
	link := filepath.Join(m.tags, name)
	os.Remove(link)
	if rel, err := filepath.Rel(m.tags, m.cwd); err == nil {
		m.catch(os.Symlink(rel, link))
	}
}
