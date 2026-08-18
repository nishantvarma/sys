func (e *editor) change() {
	a, c := e.rng()
	e.b.snapshot(e.caret)
	e.reg = e.b.cut(a, c)
	e.startInsert(a)
}

func (e *editor) changeToEnd() {
	if e.selToEnd() {
		e.change()
		return
	}
	e.b.snapshot(e.caret)
	e.startInsert(e.head())
}
