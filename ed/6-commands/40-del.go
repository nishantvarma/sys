func (e *editor) del() {
	a, c := e.rng()
	e.b.snapshot(e.caret)
	e.reg = e.b.cut(a, c)
	e.place(a)
}
