func (e *editor) change() {
	a, c := e.rng()
	e.b.snapshot()
	e.reg = e.b.cut(a, c)
	e.startInsert(a)
}
