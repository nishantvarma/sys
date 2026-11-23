func (e *editor) doSave() {
	if e.b.path == "" {
		e.msg = "no file: cur names nothing ed edits"
		return
	}
	if err := e.b.save(); err != nil {
		e.msg = "save failed: " + err.Error()
		return
	}
	e.msg = "wrote " + e.b.path
	// the one point ed knows the address is real: the file exists and the
	// line is where the work is, so cur follows what the verb made. e is
	// the verb: the log line for a save is the command that reopens it.
	// ed -f says nothing: it follows cur, which fm says, and a value has
	// one writer.
	if e.curf == "" {
		addr := outline.At(e.b.path, e.head().line+1)
		exec.Command("c", "-v", "e", addr).Run()
	}
}
