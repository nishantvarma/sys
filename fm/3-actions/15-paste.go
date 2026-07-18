func (m *fm) paste() {
	var failed, kept []string
	for _, src := range m.clip {
		dst := filepath.Join(m.cwd, filepath.Base(src))
		if exists(dst) {
			name, ok := m.prompt(
				fmt.Sprintf("name [%s]: ", filepath.Base(src)),
				nil,
				"",
			)
			if !ok {
				// cancelled: leave it pending
				kept = append(kept, src)
				continue
			}
			if name == "" {
				name = filepath.Base(src)
			}
			dst = filepath.Join(m.cwd, name)
		}
		var err error
		switch {
		case m.cutting:
			err = movePath(src, dst)
		case isDir(src):
			err = copyTree(src, dst)
		default:
			err = copyFile(src, dst)
		}
		if err != nil {
			failed = append(failed, filepath.Base(src))
			kept = append(kept, src) // failed: leave it pending
		}
	}
	if m.cutting {
		m.clip = kept // keep only what did not move
	}
	if len(failed) > 0 {
		m.flash("paste failed: " + strings.Join(failed, " "))
	}
}
