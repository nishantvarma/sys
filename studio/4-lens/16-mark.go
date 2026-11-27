// mark puts the cursor on the row at the line the lens is on, or on the
// last one above it, as ed's cursor is in the item it is in; with no row
// in that file at or above that line, the view puts the cursor on the
// first.
func mark(n *outline.Node, at string) {
	path, line := outline.Address(at)
	var cur *outline.Node
	for _, k := range n.Kids {
		if p, l := outline.Address(k.ID); p == path && l > 0 &&
			l <= line {
			cur = k
		}
	}
	if cur != nil {
		cur.Cur = true
	}
}
