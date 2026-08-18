func (e *editor) startInsert(p pos) { // caller has snapshotted
	p = e.b.clamp(p, true)
	e.caret.anchor, e.caret.head = p, p
	e.ext = false
	e.goal = p.col
	e.mode = e.ins
}
