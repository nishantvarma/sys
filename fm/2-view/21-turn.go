// turn is going in and out in the tree, the view's own. Enter and right
// open or close a folder or an item with kids in place, and right a file
// too, asking fm for what is inside; enter on a file, or on an item with
// no kids, is fm's, for it edits. Left and h close what is open, else go
// up to the node a row is in. turned is false zoomed, where in and out are
// fm's, and at the top, where going out is fm's.
func (v *term) turn(name string) (ev outline.Key, turned bool) {
	if v.lay != "tree" || v.last == nil {
		return ev, false
	}
	_, rows := v.shown()
	if len(rows) == 0 {
		return ev, false
	}
	i := curRow(rows)
	r := rows[i]
	switch name {
	case tty.KEnter, tty.KRight:
		leaf := r.n.Kind == "item" && len(r.n.Kids) == 0
		file := r.n.Kind != "dir" && r.n.Kind != "item"
		if leaf || name == tty.KEnter && file {
			return ev, false
		}
		if r.n.Open {
			return outline.Key{Name: "close", Node: r.n.ID}, true
		}
		return outline.Key{Name: "open", Node: r.n.ID}, true
	case tty.KLeft, "h":
		if r.n.Open {
			return outline.Key{Name: "close", Node: r.n.ID}, true
		}
		for j := i - 1; j >= 0; j-- {
			if rows[j].depth < r.depth {
				return outline.Key{Node: rows[j].n.ID}, true
			}
		}
		return ev, false
	}
	return ev, false
}
