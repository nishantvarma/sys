// tree is folder dir as a node, shown as text: its rows, and in each that
// is open, as a view asked or on the way down to the cursor, what opening
// it shows (see inside). The way down is always there, through a hidden
// folder or a link.
func (m *fm) tree(dir, text string) *outline.Node {
	n := &outline.Node{ID: dir, Kind: "dir", Text: text, Open: true}
	files := m.files
	if dir != m.cwd {
		files = entries(dir, m.hidden, false)
	} else if m.item == "" { // the cursor is on one of these rows
		n.In = true
	}
	way := toward(dir, m.cwd)
	hit := func(f entry) bool { return f.name == way }
	if way != "" && !slices.ContainsFunc(files, hit) { // hidden, or gone
		f := entry{name: way, path: filepath.Join(dir, way), dir: true}
		files = append(files, f)
	}
	for i, f := range files {
		k := &outline.Node{
			ID:   id(dir, f.name),
			Kind: kind(f),
			Text: f.name,
			Tags: facts(f),
			Cur:  dir == m.cwd && i == m.idx && m.item == "",
			Sel:  m.sel[f.path],
		}
		shut := f.name == "." || f.name == ".." || !m.open[f.path]
		if f.name == way || f.path == m.file() || !shut {
			m.inside(k, f.path)
		}
		n.Kids = append(n.Kids, k)
	}
	return n
}
