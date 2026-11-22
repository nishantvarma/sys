func (m *fm) rm() {
	// . and .., as rm(1) refuses them
	paths := slices.DeleteFunc(m.targets(), func(p string) bool {
		return p == m.cwd || p == filepath.Dir(m.cwd)
	})
	if len(paths) == 0 {
		return
	}
	var names []string
	for _, p := range paths {
		names = append(names, filepath.Base(p))
	}
	// the question is the status, and the next key its answer
	m.msg = "rm " + strings.Join(names, " ") + "? "
	m.draw()
	m.msg = ""
	if k, ok := m.ui.Key(); !ok || k.Name != "y" {
		return
	}
	var failed []string
	for _, p := range paths {
		rm := os.Remove
		if isRealDir(p) {
			rm = os.RemoveAll
		}
		if rm(p) != nil {
			failed = append(failed, filepath.Base(p))
		}
	}
	if len(failed) > 0 {
		m.flash("rm failed: " + strings.Join(failed, " "))
	}
	m.sel = map[string]bool{}
}
