func (e *editor) selectLine() {
	a, h := e.caret.anchor, e.caret.head
	whole := a.col == 0 && h.line >= a.line &&
		h.col == e.b.lineLen(h.line) && h.line < e.b.last()
	if whole { // whole lines already: grow down
		e.caret.head = pos{h.line + 1, e.b.lineLen(h.line+1)}
		return
	}
	l := e.head().line
	e.caret.anchor = pos{l, 0}
	e.caret.head = pos{l, e.b.lineLen(l)} // through the line break
}
