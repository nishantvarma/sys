func (e *editor) toEnd() {
	e.point(lineEnd(e.b, e.head()))
}
