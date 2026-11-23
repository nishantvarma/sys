// unfold opens every node with kids, so that the whole tree shows.
func unfold(ns []*outline.Node) {
	for _, n := range ns {
		if len(n.Kids) > 0 {
			n.Open = true
			unfold(n.Kids)
		}
	}
}
