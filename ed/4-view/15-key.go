// Key reads the next key: the terminal's, or the mouse's as ed hears it
// (mouse), a report that says nothing to ed a resize, so that the loop
// draws what it did; ctrl-v pastes X's clipboard, as a terminal's paste
// does. A left press on the bar turns to the lens it names (onBar), and
// Tab, in any mode but insert, where it types, to the next; while a lens
// other than source shows, every key is that lens's (lensKey).
func (v *term) Key() (outline.Key, bool) {
	k, ok := v.t.ReadKey()
	if !ok {
		return outline.Key{}, false
	}
	again := outline.Key{Name: tty.KResize}
	if i, on := v.onBar(k); on {
		if i >= 0 {
			v.lens = i
		}
		return again, true
	}
	if v.lens > 0 {
		return v.lensKey(k), true
	}
	switch {
	case k.Name == tty.KMouse || k.Name == tty.KMouseUp:
		if ev, said := v.mouse(k); said {
			return ev, true
		}
		return again, true
	case k.Name == "C-v":
		return outline.Key{Name: tty.KPaste,
			Text: tty.Get(tty.Clipboard)}, true
	case k.Token() == tty.KTab && v.mode != "insert":
		v.turn()
		return again, true
	}
	return outline.Key{Name: k.Token(), Text: k.Text}, true
}
