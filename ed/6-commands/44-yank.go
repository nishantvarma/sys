// yank copies the selection into reg, and mirrors it to the system
// clipboard by way of the view, which knows how its screen reaches it.
func (e *editor) yank() {
	a, c := e.rng()
	e.reg = e.b.text(a, c)
	e.ui.Copy(string(e.reg))
}
