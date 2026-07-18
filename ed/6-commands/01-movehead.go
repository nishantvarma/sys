func (e *editor) moveHead(dl, dc int) {
	h := e.head()
	e.point(pos{h.line + dl, h.col + dc})
}
