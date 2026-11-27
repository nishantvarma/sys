// weights is how a split shares its parent: each slot's share, as layout
// gives it, and all of it to a slot alone.
func weights(kids []*slot) []int {
	w := make([]int, len(kids))
	for i, k := range kids {
		w[i] = k.share
		if len(kids) == 1 {
			w[i] = 100
		}
	}
	return w
}
