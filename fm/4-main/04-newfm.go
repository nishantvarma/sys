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
	win, on := state.Pane()
	m := &fm{
		cwd:  cwd,
		this: state.Path("cur", win),
		on:   on,
		pos:  map[string]string{},
		sel:  map[string]bool{},
	}
	if name != "" {
		m.pos[cwd] = name
	}
	// in a frame's slot fm is the frame's face, and its q asks first
	// while ed holds work unsaved (quit)
	if state.Slot() != "" {
		m.unsaved = state.Path("unsaved", win)
	}
	m.bind()
	return m, nil
}
