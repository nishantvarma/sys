// paste inserts the register after the cursor; P mirrors it before,
// and either replaces an active selection.
func (e *editor) paste() { e.pasteDir(false) }
