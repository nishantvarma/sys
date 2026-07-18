func (insertMode) key(e *editor, k tty.Key) {
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
	case k.Name == "" && unicode.IsPrint(k.Rune):
		e.place(e.b.insert(e.head(), k.Rune))
	}
}
