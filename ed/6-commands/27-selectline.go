func (e *editor) selectLine() {
	a, h := e.caret.anchor, e.caret.head
	whole := a.col == 0 && h.line >= a.line &&
		h.col == len(e.b.lines[h.line]) && h.line < e.b.last()
	if whole { // whole lines already: grow down
		e.caret.head = pos{h.line + 1, len(e.b.lines[h.line+1])}
		return
	}
	l := e.head().line
	e.caret.anchor = pos{l, 0}
	e.caret.head = pos{l, len(e.b.lines[l])} // through the line break
}
