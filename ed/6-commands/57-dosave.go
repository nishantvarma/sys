func (e *editor) doSave() {
	if err := e.b.save(); err != nil {
		e.msg = "save failed: " + err.Error()
	} else {
		e.msg = "wrote " + e.b.path
	}
}
