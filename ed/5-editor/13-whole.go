// whole reports whether the selection is whole lines: the anchor at a
// line start, the head on the line break of that line or one below. A
// caret on a blank line is whole too, since the break is all that line
// is — x grows it, and only moveHead asks for a caret first.
func (e *editor) whole() bool {
	a, h := e.caret.anchor, e.caret.head
	return a.col == 0 && h.line >= a.line && h.col == e.b.lineLen(h.line)
}
