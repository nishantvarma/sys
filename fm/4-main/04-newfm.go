func newFM(path string) (*fm, error) {
	cwd, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	tags := filepath.Join(userHome, tagsRel)
	if err := os.MkdirAll(tags, 0o755); err != nil {
		return nil, err
	}
	// a file is a row and not a place: stand in its directory and land
	// on its name, the pos map's own answer for a directory seen before.
	name := ""
	if !isDir(cwd) {
		cwd, name = filepath.Dir(cwd), filepath.Base(cwd)
	}
	id, on := pane()
	m := &fm{
		cwd:  cwd,
		tags: tags,
		this: filepath.Join(userHome, curRel, id),
		on:   on,
		pos:  map[string]string{},
		sel:  map[string]bool{},
	}
	if name != "" {
		m.pos[cwd] = name
	}
	m.bind()
	return m, nil
}
