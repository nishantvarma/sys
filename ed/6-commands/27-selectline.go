func (e *editor) selectLine() {
	h := e.caret.head
	if e.whole() && h.line < e.b.last() { // whole lines already: grow down
		e.caret.head = pos{h.line + 1, e.b.lineLen(h.line+1)}
		return
	}
	l := h.line
	e.caret.anchor = pos{l, 0}
	e.caret.head = pos{l, e.b.lineLen(l)} // through the line break
}
