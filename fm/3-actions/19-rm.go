func (m *fm) rm() {
	paths := m.targets()
	if len(paths) == 0 {
		return
	}
	var names []string
	for _, p := range paths {
		names = append(names, filepath.Base(p))
	}
	_, h := m.t.Size()
	m.t.Write(tty.Status(h, "rm "+strings.Join(names, " ")+"? ", true))
	m.t.Flush()
	if k, ok := m.t.ReadKey(); !ok || k.Rune != 'y' {
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
