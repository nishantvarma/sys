func (e *editor) doSearch(forward bool) {
	if e.pat == "" {
		e.msg = "no previous search"
		return
	}
	a, c, ok := findText(e.b, e.head(), e.pat, forward)
	if !ok {
		e.msg = "not found: " + e.pat
		return
	}
	if !e.ext {
		e.caret.anchor = a
	}
	e.caret.head = c
	e.goal = c.col
}
