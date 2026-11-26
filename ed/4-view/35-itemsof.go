// itemsOf is the text n holds as the lens shown takes it, its items read
// by their indent (outline.Text) from its lines as they are: for the
// outline each item a row, two columns in a level, the item ed's cursor
// is in the view's cursor; for the render the items as they nest, which
// the view draws as their picture, parts and messages (sys/picture). Each
// item is named by its line, as ed names its lines, so a key on one is a
// key on that line.
func (v *term) itemsOf(n *outline.Node) *outline.Node {
	var b strings.Builder
	for _, l := range n.Kids {
		b.WriteString(l.Text + "\n")
	}
	items := outline.Text(strings.NewReader(b.String()), n.ID)
	tree := &outline.Node{ID: n.ID, Text: n.Text}
	if lenses[v.lens] == "render" {
		tree.Kind, tree.Kids = "picture", items
		return tree
	}
	at := curLine(n) + 1
	var cur *outline.Node
	var walk func(ns []*outline.Node, depth int)
	walk = func(ns []*outline.Node, depth int) {
		for _, it := range ns {
			row := &outline.Node{ID: it.ID, Kind: it.Kind,
				Text: strings.Repeat("  ", depth) + it.Text}
			if _, line := outline.Address(it.ID); line <= at {
				cur = row
			}
			tree.Kids = append(tree.Kids, row)
			walk(it.Kids, depth+1)
		}
	}
	walk(items, 0)
	if cur != nil {
		cur.Cur = true
	}
	return tree
}
