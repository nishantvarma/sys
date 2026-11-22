// zoom is the node the cursor is among the kids of: the folder a zoomed
// view shows. nil when no node says so.
func zoom(n *outline.Node) *outline.Node {
	if n.In {
		return n
	}
	for _, k := range n.Kids {
		if z := zoom(k); z != nil {
			return z
		}
	}
	return nil
}
