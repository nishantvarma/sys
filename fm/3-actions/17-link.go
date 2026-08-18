func (m *fm) link() {
	if len(m.clip) == 0 {
		return
	}
	for _, src := range m.clip {
		dst := filepath.Join(m.cwd, filepath.Base(src))
		if !exists(dst) {
			if rel, err := filepath.Rel(m.cwd, src); err == nil {
				m.catch(os.Symlink(rel, dst))
			}
		}
	}
	m.clip = nil
}
