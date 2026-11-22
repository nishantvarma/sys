// prompt reads a line at the status bar, through the view. ok is false when
// cancelled.
func (e *editor) prompt(msg string) (string, bool) {
	return e.ui.Ask(msg, "")
}
