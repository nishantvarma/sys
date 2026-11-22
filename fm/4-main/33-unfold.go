// unfold opens the node a view asks to see inside, when opening it shows
// anything (see inside); when not, fm says so. . and .. are places to go
// to rather than folders to show, so for them fm goes.
func (m *fm) unfold(id string) {
	if name := filepath.Base(id); name == "." || name == ".." {
		m.enter()
		return
	}
	if _, item := m.items[id]; !item && !isDir(id) {
		if _, ok := read(id); !ok {
			m.flash(filepath.Base(id) + ": not text, or too big")
			return
		}
	}
	m.open[id] = true
}
