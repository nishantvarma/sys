func header(v view) string {
	s := ""
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
		s += tty.Dim(" mark:" + filepath.Base(v.alt))
	}
	// the path is cut first, from the left: its end names where you are
	return tty.Bold(tty.ClipLeft(tilde(v.cwd), v.w-tty.Width(s))) + s
}
