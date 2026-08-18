// newLineAbove inserts a blank line above the cursor without leaving
// normal mode — Enter, unlike O, doesn't want to type. The cursor rides
// the line it pushed down, staying on its text instead of being left
// behind on the blank.
func (e *editor) newLineAbove() {
	e.b.snapshot(e.caret)
	p := e.head()
	e.b.openLine(p.line - 1)
	e.place(pos{p.line + 1, p.col})
}
