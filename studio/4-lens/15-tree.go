// tree is what the lens shows of the thing, as nodes for a view to draw:
// the thing, of the kind the lens sees it as, and under it its kids as the
// lens sees them, the one at cur's line the cursor's (mark). A lens with
// nothing to show says why, dim, in a thing of no kind.
func tree(th thing, l lens) *outline.Node {
	root := &outline.Node{ID: th.path, Text: th.path}
	kids, err := l.nodes(th)
	if err != nil {
		root.Kids = []*outline.Node{
			{ID: l.name + "#1", Kind: "waits", Text: err.Error()}}
		return root
	}
	root.Kind, root.Kids = l.kind, kids
	mark(root, th.at)
	return root
}
