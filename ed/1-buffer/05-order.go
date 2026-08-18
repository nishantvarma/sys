// order returns a, b sorted so a <= b.
func order(a, b pos) (pos, pos) {
	if b.less(a) {
		return b, a
	}
	return a, b
}
