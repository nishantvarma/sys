// recall seats a caret that came back from history — the whole selection,
// not just the head: history stored the range the edit worked on, so undo
// hands it back ready to retry. It needs no clamp. The caret was stored
// with these very lines, and clamping to the mode would eat the line break
// x rests on at col == len.
func (e *editor) recall(c sel) {
	e.caret = c
	e.ext = false
	e.goal = c.head.col
}
