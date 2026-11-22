func newFM(path string, row bool) (*fm, error) {
	cwd, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	// a file is a row and not a place, and so is cur: stand in its
	// directory and land on its name, the pos map's own answer for a
	// directory seen before.
	name := ""
	if row || !isDir(cwd) {
		cwd, name = filepath.Dir(cwd), filepath.Base(cwd)
	}
	id, on := pane()
	m := &fm{
		cwd:  cwd,
		root: cwd,
		open: map[string]bool{},
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
