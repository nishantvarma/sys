// curLine is the index of the line the cursor is in: the one with a run
// that holds it.
func curLine(n *outline.Node) int {
	for i, ln := range n.Kids {
		for _, r := range ln.Kids {
			if r.Cur {
				return i
			}
		}
	}
	return 0
}
