// layout lays the slots out in the order they come (`design/lens`). The
// frame's terminals, the first slot that keeps a shell, are its panel:
// across the bottom, 38% of the height, and hidden and shown by one key
// (panelVerb). The rest take the room above it as a golden spiral: the
// first 1/φ of it, 62%, at its right; each after it 62% of what is left,
// down and across in turn, at the top or the right of it; the last what
// remains, so the first slot never shrinks for the rest.
// It is the tree the hosts carve, tmux, gio and the page alike: the
// frame's own lines stack down, and a split's slots go the other way from
// its parent's. The panel comes back with the tree, none for a frame of no
// shell.
func layout(slots []*slot) ([]*slot, *slot) {
	var rest []*slot
	var panel *slot
	for _, s := range slots {
		if panel == nil && s.cmd == "" {
			panel = s
			continue
		}
		rest = append(rest, s)
	}
	var above *slot
	switch {
	case len(rest) == 0:
		return slots, panel
	case len(rest) == 1:
		above = rest[0]
	default:
		above = &slot{kids: spiral(rest, true)}
	}
	if panel == nil {
		above.share = 0
		return []*slot{above}, nil
	}
	above.share, panel.share = golden, 100-golden
	return []*slot{above, panel}, panel
}
