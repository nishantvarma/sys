// quit ends ed, an unsaved buffer taking a second q, which discards it.
// In a frame's slot it ends nothing, for there q is fm's, and ends the
// frame (`fm/4-main/40-quit.go`).
func (e *editor) quit() {
	if e.unsaved != "" {
		e.msg = "q in fm ends the frame"
		return
	}
	if e.b.dirty && !e.armed {
		e.armed = true
		e.msg = "modified — q again to discard"
		return
	}
	e.done = true
}
