func (e *editor) selectAll() {
	last := e.b.last()
	e.caret.anchor = pos{0, 0}
	e.caret.head = e.b.clamp(pos{last, len(e.b.lines[last])}, false)
}
