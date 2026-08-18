func (e *editor) collapse() {
	e.caret.anchor = e.caret.head
	e.ext = false
}
