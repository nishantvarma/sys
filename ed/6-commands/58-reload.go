func (e *editor) reload() {
	nb, err := load(e.b.path)
	if err != nil {
		e.msg = "reload failed: " + err.Error()
		return
	}
	e.b.snapshot() // unsaved edits stay one undo away
	e.b.lines = nb.lines
	e.b.dirty = false
	e.place(e.head())
	e.msg = "read " + e.b.path
}
