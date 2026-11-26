// lensed is the text n, as draw says it, as the lens shown takes it: its
// items read by their indent (outline.Text) from the lines as they are,
// unsaved ones too. For the outline each item is a row, two columns in a
// level, the item ed's cursor is in the cursor; for the render the items
// as they nest, which a view draws as their picture (sys/picture). Each
// item is named by its line, as draw names lines, so a key on one is a
// key on that line, and the lenses ride along, for a view's bar.
func (e *editor) lensed(n *outline.Node) *outline.Node {
	var b strings.Builder
	for _, l := range n.Kids {
		b.WriteString(l.Text + "\n")
	}
	items := outline.Text(strings.NewReader(b.String()), n.ID)
	tree := &outline.Node{ID: n.ID, Kind: "outline", Text: n.Text,
		Tags: n.Tags}
	if lenses[e.lens] == "render" {
		tree.Kind, tree.Kids = "picture", items
		return tree
	}
	at := e.head().line + 1
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
