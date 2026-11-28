// mark puts the cursor on the part at the line the lens is on, or on the
// nearest above it, as ed's cursor is in the part it is in: of the parts
// in that file, the one starting last at or above the line, the deepest
// of those that start on it, whatever order the reader put them in, as
// Go's puts a method under its type; with none, the view puts the cursor
// on the first. It takes the cursor off any part it was on, for a tree is
// marked again as the cursor moves.
func mark(n *outline.Node, at string) {
	path, line := outline.Address(at)
	var cur *outline.Node
	best := 0
	var walk func(ns []*outline.Node)
	walk = func(ns []*outline.Node) {
		for _, k := range ns {
			k.Cur = false
			if p, l := outline.Address(k.ID); p == path && l > 0 &&
				l <= line && l >= best {
				cur, best = k, l
			}
			walk(k.Kids)
		}
	}
	walk(n.Kids)
	if cur != nil {
		cur.Cur = true
	}
}
