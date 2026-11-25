// yank copies the selection into reg, and mirrors it to X's clipboard by
// way of the view, which knows how its screen reaches it, and says when
// it cannot.
func (e *editor) yank() {
	a, c := e.rng()
	e.reg = e.b.text(a, c)
	if err := e.ui.Copy(string(e.reg)); err != nil {
		e.msg = "not on the clipboard: " + err.Error()
	}
}
