func (e *editor) reload() {
	nb, err := load(e.b.path)
	if err != nil {
		e.msg = "reload failed: " + err.Error()
		return
	}
	e.b.snapshot(e.caret) // unsaved edits stay one undo away
	e.b.reset(nb)
	e.place(e.head())
	e.msg = "read " + e.b.path
}
