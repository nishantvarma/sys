// slotOf is the slot an item says: a name, then what runs there; or, with
// lines under it, its tabs, which take turns in it (tabOf). The room a
// slot takes is its place in the order (layout), so a share such as 40%
// is no part of its line.
func slotOf(n *outline.Node, named map[string]bool) (*slot, error) {
	name, cmd := split(n.Text)
	if err := fresh(n.ID, name, named); err != nil {
		return nil, err
	}
	if lead, _, _ := strings.Cut(cmd, " "); strings.HasSuffix(lead, "%") {
		return nil, fmt.Errorf("%s: %s: a slot takes its room by its"+
			" place in the order, not by a share", n.ID, name)
	}
	s := &slot{name: name, at: n.ID}
	if len(n.Kids) == 0 {
		s.cmd = cmd
		s.tabs = []*tab{{name: name, cmd: cmd, at: n.ID}}
		return s, nil
	}
	if cmd != "" {
		return nil, fmt.Errorf("%s: %s: runs nothing itself, for its"+
			" tabs do", n.ID, name)
	}
	for _, k := range n.Kids {
		tb, err := tabOf(k, named)
		if err != nil {
			return nil, err
		}
		s.tabs = append(s.tabs, tb)
	}
	s.cmd = s.tabs[0].cmd
	return s, nil
}
