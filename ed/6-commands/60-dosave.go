func (e *editor) doSave() {
	if err := e.b.save(); err != nil {
		e.msg = "save failed: " + err.Error()
		return
	}
	e.msg = "wrote " + e.b.path
	// the one point ed knows the address is real: the file exists and the
	// line is where the work is, so cur follows what the verb made. e is
	// the verb: the log line for a save is the command that reopens it.
	addr := e.b.path + ":" + strconv.Itoa(e.head().line+1)
	exec.Command("c", "-v", "e", addr).Run()
}
