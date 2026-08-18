// keep closes typed once normal mode is idle again — no insert under way,
// no count pending — and keeps it as the change . types if those keys took
// a snapshot. Mostly that is one key; a verb that enters insert runs on
// through esc, and a count rides with the key it multiplies. Keys that
// changed nothing are dropped, so a motion never displaces the change, and
// neither does . itself: it only queues keys.
func (e *editor) keep() {
	if e.mode != e.nor || e.count > 0 {
		return
	}
	if e.b.edits != e.seen {
		e.again = e.typed
	}
	e.typed, e.seen = nil, e.b.edits
}
