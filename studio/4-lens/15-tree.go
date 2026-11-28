// tree is what the lens shows of the thing, as nodes for a view to draw,
// and how it saw it: the thing, the tree the lens makes of it, its parts
// under it, the one at cur's line the cursor's (mark). A lens with nothing
// to show says why, dim, in a thing of no kind.
func tree(th thing, l lens) (*outline.Node, string) {
	root, word, err := l.nodes(th)
	if err != nil {
		why := &outline.Node{ID: l.name + "#1", Kind: "waits",
			Text: err.Error()}
		return &outline.Node{ID: th.path, Text: th.path,
			Kids: []*outline.Node{why}}, word
	}
	root.ID, root.Text = th.path, th.path
	mark(root, th.at)
	return root, word
}
