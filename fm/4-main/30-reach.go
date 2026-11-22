// reach puts the cursor on the node with the given id: a row, in its
// folder, which becomes the cursor's when a move in a tree crosses into
// another; or an item, in its text file. What the cursor leaves stays
// open, as it was shown.
func (m *fm) reach(id string) {
	for p := m.items[m.item].up; p != ""; p = m.items[p].up {
		m.open[p] = true
	}
	m.item = ""
	if at, ok := m.items[id]; ok {
		m.reach(at.file)
		m.item = id
		return
	}
	i := strings.LastIndex(id, "/")
	if i < 0 {
		return
	}
	dir, name := id[:i], id[i+1:]
	if dir == "" {
		dir = "/"
	}
	if dir != m.cwd {
		if os.Chdir(dir) != nil {
			return
		}
		for p := m.cwd; under(p, m.root) && p != m.root; {
			m.open[p] = true
			p = filepath.Dir(p)
		}
		m.pos[m.cwd] = m.curName()
		m.cwd = dir
		m.ls()
		m.title("fm:" + m.cwd)
	}
	if j := m.at(name); j >= 0 {
		m.idx = j
	}
}
