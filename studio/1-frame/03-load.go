// load reads the frame a verb is given, or studio's own, `studio/frame`,
// by its whole path: the window keeps it, for a program anywhere.
func load(args []string) (*frame, error) {
	note := filepath.Join(root(), "studio", "frame")
	if len(args) > 0 {
		note = args[0]
	}
	note, err := filepath.Abs(note)
	if err != nil {
		return nil, err
	}
	src, err := outline.File(note)
	if err != nil {
		return nil, err
	}
	f, err := read(note, outline.Text(bytes.NewReader(src), note))
	if err != nil {
		return nil, err
	}
	f.note, f.src = note, string(src)
	return f, nil
}
