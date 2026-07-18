func (m *fm) complete(s string) string {
	p := expand(s)
	dir, pre := filepath.Dir(p), filepath.Base(p)
	if isDir(p) {
		dir, pre = p, ""
	}
	ents, err := os.ReadDir(dir)
	if err != nil {
		return s
	}
	var names []string
	for _, e := range ents {
		if strings.HasPrefix(e.Name(), pre) {
			names = append(names, e.Name())
		}
	}
	switch len(names) {
	case 0:
		return s
	case 1:
		full := filepath.Join(dir, names[0])
		if isDir(full) {
			return full + "/"
		}
		return full
	default:
		return filepath.Join(dir, commonPrefix(names))
	}
}
