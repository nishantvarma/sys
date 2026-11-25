// draw says what to show: the folder fm stands in (folder), and the
// status, fm's word and its state.
func (m *fm) draw() {
	m.ui.Draw(m.folder(), m.msg, m.state())
}
