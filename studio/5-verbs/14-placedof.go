// placedOf is slots as a host reads them, each weighed among the others.
func placedOf(slots []*slot) []placed {
	out := make([]placed, len(slots))
	for i, w := range weights(slots) {
		s := slots[i]
		out[i] = placed{s.name, w, s.cmd, placedOf(s.kids)}
	}
	return out
}
