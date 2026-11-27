// see is what the lens is on at the address at (thing).
func (l *lensing) see(at string) thing {
	path, _ := outline.Address(at)
	th := thing{at: at, path: path}
	note, _, ok := outline.Note(at)
	if !ok {
		th.why = errors.New("no note to read here")
		if fi, err := os.Stat(path); err != nil {
			th.why = err
		} else if fi.IsDir() {
			th.why = errors.New("a folder with no README")
		} else if _, err := outline.File(path); err != nil {
			th.why = err
		}
		return th
	}
	if b, ok := l.noted(note); ok {
		th.note, th.text, th.held = note, b, true
		return th
	}
	b, err := outline.File(note)
	if err != nil {
		th.why = err
		return th
	}
	th.note, th.text = note, b
	return th
}
