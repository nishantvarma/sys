// point moves the head; a set ext pins the anchor so the motion grows
// or shrinks the selection instead of collapsing it. The head may
// cross the anchor — rng orders them back.
func (e *editor) point(p pos) {
	if !e.ext {
		e.place(p)
		return
	}
	e.caret.head = e.b.clamp(p, false)
	e.goal = e.caret.head.col
}
