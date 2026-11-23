// sum adds a split's weights.
func sum(w []int) int {
	n := 0
	for _, x := range w {
		n += x
	}
	return n
}
