func (e *editor) deleteToEnd() {
	if e.selToEnd() {
		e.del()
	}
}
