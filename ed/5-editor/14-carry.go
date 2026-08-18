// carry moves whole lines dl lines down, or up for a negative dl, as far
// as the buffer allows. The block keeps its height, and its ends go back
// on the line start and the line break, so it is still whole lines.
func (e *editor) carry(dl int) {
	a, h := e.caret.anchor, e.caret.head
	dl = max(-a.line, min(dl, e.b.last()-h.line))
	e.caret.anchor = pos{a.line + dl, 0}
	e.caret.head = pos{h.line + dl, e.b.lineLen(h.line + dl)}
}
