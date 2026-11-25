// reach puts the cursor on the row a view's move or a click names, by its
// id, in the folder fm stands in. The folder's own id, which the head
// names as thumbs, is where fm is already, and a folder above it, which a
// part of the head's path names, is the way up there. false when the row
// is not there to reach.
func (m *fm) reach(node string) bool {
	if node == m.cwd {
		return true
	}
	if node == "/" || strings.HasPrefix(m.cwd, node+"/") {
		m.cd(node)
		return true
	}
	for i, f := range m.files {
		if id(m.cwd, f.name) == node {
			m.idx = i
			return true
		}
	}
	return false
}
