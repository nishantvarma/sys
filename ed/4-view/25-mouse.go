// mouse is a mouse report as ed hears it, and whether it says anything to
// ed. A left press is a key on the line and the rune it fell on, which
// puts the cursor there, or on the status line's mode its switch
// (modeWord); a move with it held is a drag to the same; letting go of a
// drag has its selection put on X's primary (Draw). A second press at the
// same cell soon after is a double tap, which, as a right click, hands
// the word there to p, as either does in any pane
// (`tmux/conf:/MouseDown3Pane/`); a middle click pastes the primary
// selection where it falls; and the wheel is three lines up or down.
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
	id, text := v.last.Kids[i].ID, []rune(v.last.Kids[i].Text)
	switch k.Button {
	case tty.MLeft:
		tap := k.X == v.px && k.Y == v.py &&
			time.Since(v.press) < 400*time.Millisecond
		v.press, v.px, v.py, v.dragged = time.Now(), k.X, k.Y, false
		if tap {
			plumb(wordAt(text, at-1))
			return outline.Key{}, false
		}
		return outline.Key{Node: id, At: at}, true
	case tty.MDrag:
		v.dragged = true
		return outline.Key{Name: tty.KDrag, Node: id, At: at}, true
	case tty.MMiddle:
		return outline.Key{Name: tty.KPaste, Node: id, At: at,
			Text: tty.Get(tty.Primary)}, true
	case tty.MRight:
		plumb(wordAt(text, at-1))
	}
	return outline.Key{}, false
}
