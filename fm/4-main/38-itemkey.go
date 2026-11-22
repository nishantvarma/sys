// itemKey does a key said on an item with an item's verbs, never a file's,
// so no file verb reaches the file through one of its items: e edits at
// its line; right goes in, to its first kid; enter goes in too, and on an
// item with no kids edits; left and h go out, to the node above. q, ? and
// . are fm's everywhere. It reports whether the key was the item's.
func (m *fm) itemKey(k string) bool {
	at := m.items[m.item]
	edit := func() {
		line := m.item[strings.LastIndex(m.item, ":")+1:]
		m.spawn(true, false, cmdEdit, "+"+line)
	}
	switch k {
	case "", "q", "?", ".":
		return false
	case tty.KRight, tty.KEnter:
		if at.first != "" {
			m.reach(at.first)
		} else if k == tty.KEnter {
			edit()
		}
	case "e":
		edit()
	case tty.KLeft, "h":
		m.reach(at.up)
	default:
		m.flash(k + ": not on an item")
	}
	return true
}
