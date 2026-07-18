func newFM(path string) (*fm, error) {
	cwd, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	tags := filepath.Join(userHome, tagsRel)
	if err := os.MkdirAll(tags, 0o755); err != nil {
		return nil, err
	}
	m := &fm{
		cwd:  cwd,
		tags: tags,
		pos:  map[string]int{},
		sel:  map[string]bool{},
	}
	m.bind()
	return m, nil
}
