func (e *editor) selectAll() {
	last := e.b.last()
	e.caret.anchor = pos{0, 0}
	e.caret.head = e.b.clamp(pos{last, e.b.lineLen(last)}, false)
}
