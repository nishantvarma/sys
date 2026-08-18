func (e *editor) place(p pos) {
	p = e.b.clamp(p, e.mode.pastEnd())
	e.caret.anchor, e.caret.head = p, p
	e.ext = false
	e.goal = p.col
}
