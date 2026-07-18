// clone mirrors the current file under ~/Unsynced, preserving its path
// relative to the nearest of ~/Data, ~, or /.
func (m *fm) clone(dirs bool) {
	c := m.cur()
	if c == "" || isSymlink(c) {
		return
	}
	src := c
	if r, err := filepath.EvalSymlinks(c); err == nil {
		src = r
	}
	src, _ = filepath.Abs(src)
	root := "/"
	roots := []string{filepath.Join(userHome, "Data"), userHome, "/"}
	for _, r := range roots {
		if under(src, r) {
			root = r
			break
		}
	}
	rel, _ := filepath.Rel(root, src)
	dst := filepath.Join(userHome, "Unsynced", rel)
	if !exists(dst) {
		os.MkdirAll(filepath.Dir(dst), 0o755)
		switch {
		case isFile(src):
			copyFile(src, dst)
		case dirs:
			copyTree(src, dst)
		default:
			os.Mkdir(dst, 0o755)
		}
	}
	if isDir(dst) {
		m.cd(dst)
	} else {
		m.cd(filepath.Dir(dst))
	}
}
