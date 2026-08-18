func (e *editor) deleteLines() {
	a, c := e.rng()
	e.caret.anchor = pos{a.line, 0}
	e.caret.head = pos{c.line, e.b.lineLen(c.line)}
	e.del()
}
