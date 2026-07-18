func header(v view) string {
	s := tty.Bold(tilde(v.cwd))
	if n := len(v.sel); n > 0 {
		s += tty.Dim(fmt.Sprintf(" [%d]", n))
	}
	if v.clip > 0 {
		mode := "cp"
		if v.cutting {
			mode = "cut"
		}
		s += tty.Dim(fmt.Sprintf(" %s:%d", mode, v.clip))
	}
	if v.alt != "" {
		s += tty.Dim(" alt:" + filepath.Base(v.alt))
	}
	return s
}
