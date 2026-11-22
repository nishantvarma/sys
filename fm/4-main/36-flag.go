// flag marks the items under n with what fm holds of them: the cursor, the
// node it is among the kids of, and which are open, as a view asked or on
// the way down to the cursor. It keeps where each one is, so that a key
// said on one finds it, and reports whether the cursor is under n.
func (m *fm) flag(n *outline.Node, file string) bool {
	here := false
	for _, k := range n.Kids {
		first := ""
		if len(k.Kids) > 0 {
			first = k.Kids[0].ID
		}
		m.items[k.ID] = place{file, n.ID, first}
		if k.ID == m.item {
			k.Cur, n.In, here = true, true, true
		}
		if m.flag(k, file) {
			k.Open, here = true, true
		}
		k.Open = k.Open || m.open[k.ID]
	}
	return here
}
