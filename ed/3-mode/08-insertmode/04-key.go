// key types k at the caret. Insert is a plain editor's: the chords work
// here as in normal, and a selection, as a drag leaves one, is typed over,
// what writes taking it away first, and backspace taking only it.
func (insertMode) key(e *editor, k tty.Key) {
	e.armed = false
	if act, ok := chordMap[k.Token()]; ok {
		if act != nil {
			act(e)
		}
		return
	}
	writes := k.Name == tty.KEnter || k.Name == tty.KBack ||
		k.Name == tty.KTab || k.Name == "" && unicode.IsPrint(k.Rune)
	if writes {
		e.pay()
	}
	if a, c := e.rng(); a != c && writes {
		e.b.cut(a, c)
		e.place(a)
		if k.Name == tty.KBack {
			return
		}
	}
	switch {
	case k.Name == tty.KEsc:
		e.leaveInsert()
	case k.Name == tty.KEnter:
		e.place(e.b.split(e.head()))
	case k.Name == tty.KBack:
		e.place(e.b.backspace(e.head()))
	case k.Name == tty.KTab:
		e.place(e.b.insert(e.head(), '\t'))
	case k.Name == tty.KLeft:
		e.left()
	case k.Name == tty.KRight:
		e.right()
	case k.Name == tty.KUp:
		e.up()
	case k.Name == tty.KDown:
		e.down()
	case k.Name == tty.KPgUp:
		e.page(-1)
	case k.Name == tty.KPgDn:
		e.page(1)
	case writes:
		e.place(e.b.insert(e.head(), k.Rune))
	}
}
