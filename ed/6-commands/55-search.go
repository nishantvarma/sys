func (e *editor) search() {
	pat, ok := e.prompt("/")
	if !ok || pat == "" {
		return
	}
	e.pat = pat
	e.searchNext()
}
