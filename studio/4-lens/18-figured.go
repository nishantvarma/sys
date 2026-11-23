// figured is a note as its figure's view takes it: all of its items, which
// the view lays out as parts and messages (sys/figure) and draws, in runes
// on the terminal. A folder is no figure.
func figured(path string) ([]*outline.Node, error) {
	if fi, err := os.Stat(path); err == nil && fi.IsDir() {
		return nil, errors.New("a folder is no figure; a note in it" +
			" may be")
	}
	b, err := outline.File(path)
	if err != nil {
		return nil, err
	}
	return outline.Text(bytes.NewReader(b), path), nil
}
