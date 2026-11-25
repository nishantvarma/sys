// cutSel is ctrl-x: d, and what it took onto the clipboard as well as into
// the register; in insert only a selection, as copySel.
func (e *editor) cutSel() {
	a, c := e.rng()
	switch {
	case e.mode != e.ins:
		e.del()
	case a == c:
		return
	default:
		e.pay()
		e.reg = e.b.cut(a, c)
		e.place(a)
	}
	if err := e.ui.Copy(string(e.reg)); err != nil {
		e.msg = "not on the clipboard: " + err.Error()
	}
}
