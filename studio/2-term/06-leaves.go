// leaves is every slot that holds a pane of its own, in the frame's
// order: the slots that are not split.
func leaves(slots []*slot) []*slot {
	var out []*slot
	for _, s := range slots {
		if len(s.kids) == 0 {
			out = append(out, s)
			continue
		}
		out = append(out, leaves(s.kids)...)
	}
	return out
}
