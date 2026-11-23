// mark puts the cursor on the node at the address the lens is on, and
// says its parent holds it, as fm marks a row in its folder; with no node
// there, the view puts the cursor on the first.
func mark(n *outline.Node, at string) bool {
	for _, k := range n.Kids {
		if k.ID == at {
			k.Cur, n.In = true, true
			return true
		}
		if mark(k, at) {
			return true
		}
	}
	return false
}
