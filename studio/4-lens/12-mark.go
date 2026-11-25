// mark puts the cursor on the row at the address the lens is on, as fm
// marks its row; with no row there, the view puts the cursor on the
// first.
func mark(n *outline.Node, at string) {
	for _, k := range n.Kids {
		if k.ID == at {
			k.Cur = true
			return
		}
	}
}
