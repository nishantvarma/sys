// draw says what to show, as an outline: the tree from fm's top, every
// open folder with its rows and open text file with its items, and the
// cursor's place; and the status, fm's word and its state. How it looks,
// laid out and where, is the view's. A cursor on an item that is gone, its
// file changed under it, goes back to the file's row.
func (m *fm) draw() {
	m.items = map[string]place{}
	t := m.tree(m.root, m.root)
	if _, ok := m.items[m.item]; m.item != "" && !ok {
		m.item = ""
		m.items = map[string]place{}
		t = m.tree(m.root, m.root)
	}
	m.ui.Draw(t, m.msg, m.state())
}
