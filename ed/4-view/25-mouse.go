// mouse is a mouse report as ed hears it, and whether it says anything to
// ed, the three buttons as in any pane (`design/mouse`). A left press is a
// key on the line and the rune it fell on, which puts the cursor there,
// or on the status line's mode its switch (modeWord); a move with it held
// is a drag to the same; letting go of a drag has its selection put on
// X's primary (Draw). A right click hands what it fell on
// (outline.Clicked) to p, which opens what it names, and so does a second
// press at the same cell soon after, a double tap, the touchpad's right
// click; a middle click hands it to the shell, at its prompt (p -a); and
// the wheel is three lines up or down.
func (v *term) mouse(k tty.Key) (outline.Key, bool) {
	if k.Name == tty.KMouseUp {
		v.ended, v.dragged = v.dragged, false
		return outline.Key{}, false
	}
	switch k.Button {
	case tty.MWheelUp, tty.MWheelDown:
		name := tty.KUp
		if k.Button == tty.MWheelDown {
			name = tty.KDown
		}
		v.t.Feed([]tty.Key{{Name: name}, {Name: name}})
		return outline.Key{Name: name}, true
	case tty.MLeft:
		if name := v.modeWord(k.X, k.Y); name != "" {
			return outline.Key{Name: name}, true
		}
	}
	i, at := v.at(k.X, k.Y)
	if i < 0 {
		return outline.Key{}, false
	}
	id := v.last.Kids[i].ID
	switch k.Button {
	case tty.MLeft:
		tap := k.X == v.px && k.Y == v.py &&
			time.Since(v.press) < 400*time.Millisecond
		v.press, v.px, v.py, v.dragged = time.Now(), k.X, k.Y, false
		if tap {
			plumb(outline.Clicked(v.last, i, at))
			return outline.Key{}, false
		}
		return outline.Key{Node: id, At: at}, true
	case tty.MDrag:
		v.dragged = true
		return outline.Key{Name: tty.KDrag, Node: id, At: at}, true
	case tty.MMiddle:
		plumb(outline.Clicked(v.last, i, at), "-a")
	case tty.MRight:
		plumb(outline.Clicked(v.last, i, at))
	}
	return outline.Key{}, false
}
