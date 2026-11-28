// draw says what to show: the folder fm stands in (folder), and the
// status, fm's word and its state; but not the same again, as the lens
// does not. A key is news, so a draw follows each; a wake is news only
// when it moved the cursor, read the folder again or resized the view,
// and fm's own word to cur, coming back, does none of those.
func (m *fm) draw() {
	w, h := m.ui.Size()
	d := fmt.Sprint(m.keyed, m.reads, m.cwd, m.idx, m.msg, m.state(), w,
		h)
	if d == m.drew {
		return
	}
	m.drew = d
	m.ui.Draw(m.folder(), m.msg, m.state())
}
