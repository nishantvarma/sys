// prompt reads a line at the status bar. ok is false when cancelled.
func (e *editor) prompt(msg string) (string, bool) {
	return tty.ReadLine(e.t, msg, nil, "")
}
