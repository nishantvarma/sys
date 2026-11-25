// slotOf is the slot an item says, and the slots under it: a name, then
// a share such as 40%, then what runs there. A name is one word (word),
// and none of the names studio keeps its own by beside the slots', frame,
// here and asked (open, closeVerb); no two slots share one, for a name is
// how a program finds a slot. A slot split into others runs nothing
// itself.
func slotOf(n *outline.Node, named map[string]bool) (*slot, error) {
	name, rest := split(n.Text)
	pct, cmd, ok := share(rest)
	why := ""
	switch {
	case !word(name):
		why = "a slot's name is one word: letters, digits, - and _"
	case name == "frame" || name == "here" || name == "asked":
		why = "studio keeps its own by that name"
	case named[name]:
		why = "a second slot by that name"
	case !ok:
		why = "a share is a whole percent, from 1% to 100%"
	}
	if why != "" {
		return nil, fmt.Errorf("%s: %s: %s", n.ID, name, why)
	}
	named[name] = true
	s := &slot{name: name, share: pct, cmd: cmd, at: n.ID}
	for _, k := range n.Kids {
		ks, err := slotOf(k, named)
		if err != nil {
			return nil, err
		}
		s.kids = append(s.kids, ks)
	}
	if len(s.kids) > 0 && s.cmd != "" {
		return nil, fmt.Errorf("%s: %s: runs nothing, for it is split",
			n.ID, name)
	}
	return s, room(s.kids, n.ID)
}
