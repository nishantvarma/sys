// copySel is ctrl-c: y, but in insert only a selection, for a caret there
// sits between runes and holds none.
func (e *editor) copySel() {
	if a, c := e.rng(); e.mode == e.ins && a == c {
		return
	}
	e.yank()
}
