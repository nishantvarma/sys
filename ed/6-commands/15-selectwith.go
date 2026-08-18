// selectWith extends the selection along a motion; a set ext keeps the
// standing anchor instead of re-anchoring at the cursor. The head may
// rest on a line break (col == len), like selectLine's, so a word hop
// off a line's end selects through it and wraps.
func (e *editor) selectWith(m func(*buffer, pos) pos) {
	from := e.head()
	if !e.ext {
		e.caret.anchor = from
	}
	e.caret.head = e.b.clamp(m(e.b, from), true)
	e.goal = e.caret.head.col
}
