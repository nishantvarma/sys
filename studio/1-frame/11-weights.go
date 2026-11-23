// weights is how a split shares its parent: a slot's own share, or an
// even part of what the given shares leave, never less than one.
func weights(kids []*slot) []int {
	given, none := tally(kids)
	w := make([]int, len(kids))
	for i, k := range kids {
		w[i] = k.share
		if w[i] == 0 {
			w[i] = max(1, (100-given)/none)
		}
	}
	return w
}
