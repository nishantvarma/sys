func (m *fm) paste() {
	var failed, kept []string
	made := ""
	for _, src := range m.clip {
		dst := filepath.Join(m.cwd, filepath.Base(src))
		if exists(dst) {
			// a clash: a name typed is the act, the taken one
			// typed again to write over it, for here a default
			// would destroy; nothing typed leaves it pending
			name, ok := m.prompt(filepath.Base(src)+
				" is taken, name: ", "")
			if !ok || name == "" {
				kept = append(kept, src)
				continue
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
		} else {
			made = filepath.Base(dst)
		}
	}
	if m.cutting {
		m.clip = kept // keep only what did not move
	}
	if made != "" {
		m.ls() // land on what arrived, the last of it
		m.land(made)
	}
	if len(failed) > 0 {
		m.flash("paste failed: " + strings.Join(failed, " "))
	}
}
