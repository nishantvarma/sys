// tabOf is the tab an item under a slot says: a name, then what runs in
// it, nothing for a shell. A tab is one program, so nothing goes under it.
func tabOf(n *outline.Node, named map[string]bool) (*tab, error) {
	name, cmd := split(n.Text)
	if err := fresh(n.ID, name, named); err != nil {
		return nil, err
	}
	if len(n.Kids) > 0 {
		return nil, fmt.Errorf("%s: %s: a tab is one program, and"+
			" nothing goes under it", n.ID, name)
	}
	return &tab{name: name, cmd: cmd, at: n.ID}, nil
}
