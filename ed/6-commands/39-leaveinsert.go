func (e *editor) leaveInsert() {
	e.mode = e.nor
	h := e.head()
	e.place(
		pos{h.line, h.col - 1},
	) // step onto the last typed rune, Vim-style
}
