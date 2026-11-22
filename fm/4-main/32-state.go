// state is what fm says of itself on the status line, only what is not as
// it starts: how many are selected, the clip and whether it cuts, the
// mark, dotfiles when they show, and the search n and N go on with.
func (m *fm) state() []string {
	var t []string
	if k := len(m.sel); k > 0 {
		t = append(t, fmt.Sprintf("[%d]", k))
	}
	if c := len(m.clip); c > 0 {
		mode := "cp"
		if m.cutting {
			mode = "cut"
		}
		t = append(t, fmt.Sprintf("%s:%d", mode, c))
	}
	if m.alt != "" {
		t = append(t, "mark:"+filepath.Base(m.alt))
	}
	if m.hidden {
		t = append(t, "dotfiles")
	}
	if m.pat != "" {
		t = append(t, "/"+m.pat)
	}
	return t
}
