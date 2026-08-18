// moveHead moves the head dl lines and dc columns, for j k and the page
// keys. Whole lines are carried instead, unless v is on: they have no
// column to aim, so the move is all in dl. A caret is never carried, not
// even on a blank line where it is the line too, so j still walks one
// through blank lines.
func (e *editor) moveHead(dl, dc int) {
	if !e.ext && e.caret.anchor != e.caret.head && e.whole() {
		e.carry(dl)
		return
	}
	h := e.head()
	e.point(pos{h.line + dl, h.col + dc})
}
