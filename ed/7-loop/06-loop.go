// loop reads a key at a time, the keys . put back first, else the view's,
// and hands each to the mode, until ed is done or the view is gone; in a
// frame it marks, as it goes, what it holds unsaved (mark).
func (e *editor) loop() {
	defer e.mark(false)
	for !e.done {
		if e.curf != "" {
			e.chase()
		}
		e.mark(e.b.dirty)
		// Only the last frame of a burst is worth painting, and a
		// paste is a long burst: draw once the queue has run dry.
		if len(e.fed) == 0 && (e.t == nil || !e.t.Pending()) {
			e.draw()
		}
		ev, ok := e.next()
		if !ok {
			break
		}
		// a press on a lens's name, on a view's bar, shows that lens
		if ev.Name == tty.KLens {
			e.toLens(ev.Text)
			continue
		}
		// a key on a line, as a click or a lens says one, puts the
		// cursor there first, and a landing alone is nothing more; a
		// drag moves the head alone; a paste lands where it falls,
		// past a line's end too. One on a line not ed's goes with it,
		// for it was meant for that line and not the cursor's. A click
		// in an insert is no key . types again.
		if ev.Node != "" {
			p, ok := e.spot(ev.Node, ev.At)
			switch {
			case !ok:
				e.msg = "gone: " + ev.Node
				continue
			case ev.Name == tty.KDrag:
				e.drag(p)
				continue
			case ev.Name == tty.KPaste:
				q := e.b.clamp(p, true)
				e.caret, e.ext = sel{q, q}, false
			default:
				e.place(p)
				e.unkept = e.unkept || e.mode == e.ins
			}
			// in a lens, enter on an item, a right click's too, is
			// source on its line, and no key of the text's
			if e.lens > 0 && ev.Name == tty.KEnter {
				e.lens = 0
				continue
			}
			if ev.Name == "" {
				continue
			}
		}
		k := keyOf(ev)
		if k.Name == tty.KResize { // redraw at the top of the loop
			continue
		}
		if e.lens > 0 { // nothing is typed in a lens but source
			e.lensKey(k)
			continue
		}
		e.msg = ""
		e.typed = append(e.typed, k)
		// A paste is text, not keystrokes: it goes in whole, in
		// either mode, and never through a binding.
		if k.Name == tty.KPaste {
			e.pasteText(k.Text)
		} else {
			e.mode.key(e, k)
		}
		e.keep()
	}
}
