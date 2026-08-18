func (e *editor) change() {
	a, c := e.rng()
	e.b.snapshot(e.caret)
	e.reg = e.b.cut(a, c)
	e.startInsert(a)
}
