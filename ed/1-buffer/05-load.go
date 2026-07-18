func load(path string) (*buffer, error) {
	b := &buffer{path: path, lines: [][]rune{{}}}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return b, nil // new file
		}
		return nil, err
	}
	b.lines = nil
	for _, ln := range strings.Split(string(data), "\n") {
		b.lines = append(b.lines, []rune(ln))
	}
	// A trailing newline yields a spurious empty last line; drop it.
	if n := len(b.lines); n > 1 && len(b.lines[n-1]) == 0 {
		b.lines = b.lines[:n-1]
	}
	if len(b.lines) == 0 {
		b.lines = [][]rune{{}}
	}
	return b, nil
}
