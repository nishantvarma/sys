// loop draws, takes a key and acts on it, until a quit or the view is gone.
func (m *fm) loop() {
	for !m.done {
		m.ls()
		m.say()
		m.draw()
		ev, ok := m.ui.Key()
		k := ev.Name
		if !ok || k == tty.KInt {
			return
		}
		if k == tty.KResize { // redraw at the top of the loop
			continue
		}
		if k == tty.KFocus || k == tty.KBlur {
			m.on = k == tty.KFocus
			// a still row is silent, the eye coming back is
			// not: another pane spoke while we were away.
			m.said = ""
			continue
		}
		// a key on a row, as a click or a move, puts the cursor there
		if ev.Node != "" {
			m.reach(ev.Node)
		}
		m.msg = ""
		switch b, hit := m.keys[k]; {
		case k == "open": // a view asks to see inside a node
			m.unfold(ev.Node)
		case k == "close":
			delete(m.open, ev.Node)
		case m.item != "" && m.itemKey(k): // on an item, its verbs
		case hit && b.act != nil:
			b.act()
		}
	}
}
