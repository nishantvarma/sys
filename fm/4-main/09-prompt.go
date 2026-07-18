// prompt reads a line at the status bar. ok is false when cancelled.
func (m *fm) prompt(
	msg string,
	complete func(string) string,
	def string,
) (string, bool) {
	return tty.ReadLine(m.t, msg, complete, def)
}
