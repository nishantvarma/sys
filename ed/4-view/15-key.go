// Key reads the next key: the terminal's, or the mouse's as ed hears it
// (mouse), a report that says nothing to ed a resize, so that the loop
// draws what it did; ctrl-v pastes X's clipboard, as a terminal's paste
// does. A left press on the bar names the lens there (onBar). While ed
// shows a lens other than source, sys/view takes each key and click
// first, as it does fm's, and hands on the rest (design/lens).
func (v *term) Key() (outline.Key, bool) {
	k, ok := v.t.ReadKey()
	if !ok {
		return outline.Key{}, false
	}
	again := outline.Key{Name: tty.KResize}
	if name, on := v.onBar(k); on {
		if name == "" {
			return again, true
		}
		return outline.Key{Name: tty.KLens, Text: name}, true
	}
	if v.last != nil && v.last.Kind != "text" && v.seen != nil {
		if ev, theirs := v.seen.Handle(k); theirs {
			return ev, true
		}
		return again, true
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
	}
	return outline.Key{Name: k.Token(), Text: k.Text}, true
}
