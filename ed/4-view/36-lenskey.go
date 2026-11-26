// lensKey is what k does while a lens other than source shows, where
// nothing is typed: Tab turns to the next lens, esc goes back to source,
// and enter goes there on the line of the item the view's cursor is on;
// ctrl-s saves, as in every mode. The rest are the view's: a move, or a
// left click on an item, is a key on its line, which puts ed's cursor
// there and so keeps the view's on that item; a right click or a double
// tap on one goes to source on its line; t lays the outline out as rows
// or as thumbs, the picture scrolls, and a middle click hands an item to
// the shell. Any other key is nothing here, a resize, which draws again.
func (v *term) lensKey(k tty.Key) outline.Key {
	again := outline.Key{Name: tty.KResize}
	switch tok := k.Token(); tok {
	case tty.KTab:
		v.turn()
		return again
	case tty.KEsc:
		v.lens = 0
		return again
	case "C-s":
		return outline.Key{Name: tok}
	case tty.KEnter:
		v.lens = 0
		if c := v.cursor(); c != "" {
			return outline.Key{Node: c}
		}
		return again
	}
	if v.seen == nil {
		return again
	}
	ev, theirs := v.seen.Handle(k)
	if !theirs || !v.mine(ev.Node) {
		return again
	}
	// the view's cursor goes with the key, for in a burst of keys ed
	// draws only after the last, and the next move starts from here
	for _, r := range v.items.Kids {
		r.Cur = r.ID == ev.Node
	}
	if ev.Name == tty.KEnter {
		v.lens = 0
	}
	return outline.Key{Node: ev.Node}
}

// mine is whether id names a line of the text shown, as an item does, and
// not a folder, as the parts of the view's head do.
func (v *term) mine(id string) bool {
	p, line := outline.Address(id)
	return line > 0 && v.last != nil && p == v.last.ID
}

// cursor is the item the view's cursor is on, as the outline drew it
// last; none in the picture, which has no cursor.
func (v *term) cursor() string {
	if v.items == nil || v.items.Kind == "picture" {
		return ""
	}
	for _, r := range v.items.Kids {
		if r.Cur {
			return r.ID
		}
	}
	return ""
}
