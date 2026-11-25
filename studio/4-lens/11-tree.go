// tree is what the lens shows of the thing at at, as nodes for a view to
// draw: the thing, of the kind the lens sees it as, and under it its kids
// as the lens sees them, the one at at the cursor's. A placeholder's kids
// are the words it says; a lens with nothing to show says why, dim, in a
// thing of no kind.
func tree(at string, l lens) *outline.Node {
	path, _ := outline.Address(at)
	root := &outline.Node{ID: path, Text: path}
	words := l.waits
	if l.nodes != nil {
		kids, err := l.nodes(path)
		if err == nil {
			root.Kind, root.Kids = l.kind, kids
			mark(root, at)
			return root
		}
		words = []string{err.Error()}
	}
	for i, s := range words {
		id := l.name + "#" + strconv.Itoa(i+1)
		root.Kids = append(root.Kids,
			&outline.Node{ID: id, Kind: "waits", Text: s})
	}
	return root
}
