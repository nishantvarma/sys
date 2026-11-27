// placedOf is slots as a host reads them, each weighed among the others;
// panel is the frame's panel (shell), which says so.
func placedOf(slots []*slot, panel *slot) []placed {
	out := make([]placed, len(slots))
	for i, w := range weights(slots) {
		s := slots[i]
		out[i] = placed{Name: s.name, Weight: w, Run: s.cmd,
			Kids: placedOf(s.kids, panel), Panel: s == panel}
		if len(s.tabs) > 1 {
			for _, tb := range s.tabs {
				out[i].Tabs = append(out[i].Tabs,
					placed{Name: tb.name, Run: tb.cmd})
			}
		}
	}
	return out
}
