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
		// a key on a row, as a click or a move, puts the cursor there;
		// one on a node gone since it was drawn goes with it, for it
		// was meant for that node and not the cursor's
		if ev.Node != "" && !m.reach(ev.Node) {
			m.flash("gone: " + ev.Node)
			continue
		}
		m.msg = ""
		if b, hit := m.keys[k]; hit && b.act != nil {
			b.act()
		}
	}
}
