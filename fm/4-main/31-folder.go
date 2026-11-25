// folder is the folder fm stands in as a node, and its rows: each its
// kind, its name and its facts, and the cursor's marked. How it looks,
// as rows or as thumbs, is the view's.
func (m *fm) folder() *outline.Node {
	n := &outline.Node{ID: m.cwd, Kind: "dir", Text: m.cwd}
	for i, f := range m.files {
		n.Kids = append(n.Kids, &outline.Node{
			ID:   id(m.cwd, f.name),
			Kind: kind(f),
			Text: f.name,
			Tags: facts(f),
			Cur:  i == m.idx,
			Sel:  m.sel[f.path],
		})
	}
	return n
}
