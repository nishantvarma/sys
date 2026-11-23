// reach puts the cursor on the node with the given id: a row, in its
// folder, which becomes the cursor's when a move in a tree crosses into
// another; or an item, in its text file. What the cursor leaves stays
// open, as it was shown. false when the node is not there to reach.
func (m *fm) reach(id string) bool {
	for p := m.items[m.item].up; p != ""; p = m.items[p].up {
		m.open[p] = true
	}
	m.item = ""
	if at, ok := m.items[id]; ok {
		if !m.reach(at.file) {
			return false
		}
		m.item = id
		return true
	}
	i := strings.LastIndex(id, "/")
	if i < 0 {
		return false
	}
	dir, name := id[:i], id[i+1:]
	if dir == "" {
		dir = "/"
	}
	if dir != m.cwd {
		if os.Chdir(dir) != nil {
			return false
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
	j := m.at(name)
	if j < 0 {
		return false
	}
	m.idx = j
	return true
}
