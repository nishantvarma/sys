// prompt reads a line at the status bar, through the view. ok is false when
// cancelled.
func (m *fm) prompt(msg, def string) (string, bool) {
	return m.ui.Ask(msg, def)
}
