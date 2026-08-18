func (e *editor) reload() {
	nb, err := load(e.b.path)
	if err != nil {
		e.msg = "reload failed: " + err.Error()
		return
	}
	e.b.reset(nb, e.caret) // unsaved edits stay one undo away
	e.place(e.head())
	e.msg = "read " + e.b.path
}
