func (e *editor) search() {
	pat, ok := tty.ReadLine(e.t, "/", nil, "")
	if !ok || pat == "" {
		return
	}
	e.pat = pat
	e.searchNext()
}
