// leaveInsert goes back to normal: a caret steps onto the last typed
// rune, Vim-style; a selection, as a drag leaves one, stays for the verbs.
func (e *editor) leaveInsert() {
	e.mode, e.owed = e.nor, false
	if e.caret.anchor != e.caret.head {
		return
	}
	h := e.head()
	e.place(pos{h.line, h.col - 1})
}
