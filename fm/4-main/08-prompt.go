// prompt reads a line at the status bar. ok is false when cancelled. The
// completion seam stays tty's; fm's last completer left with goto.
func (m *fm) prompt(msg, def string) (string, bool) {
	return tty.ReadLine(m.t, msg, nil, def)
}
