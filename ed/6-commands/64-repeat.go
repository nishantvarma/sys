// repeat types the last change again: its keys go back on the input, ahead
// of anything typed, and run as they ran the first time — on whatever is
// selected now. The selection was chosen before the verb and was never
// part of the change, so it needs no motion of its own: n . is the next
// match, same change.
func (e *editor) repeat() {
	if e.again == nil {
		e.msg = "nothing to repeat"
		return
	}
	e.t.Feed(e.again)
}
