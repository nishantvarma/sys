// loop reads a key at a time, the keys . put back first, else the view's,
// and hands each to the mode, until ed is done or the view is gone.
func (e *editor) loop() {
	for !e.done {
		if e.curf != "" {
			e.chase()
		}
		// Only the last frame of a burst is worth painting, and a
		// paste is a long burst: draw once the queue has run dry.
		if len(e.fed) == 0 && (e.t == nil || !e.t.Pending()) {
			e.draw()
		}
		ev, ok := e.next()
		if !ok {
			break
		}
		// a key on a line, as a click or a lens says one, puts the
		// cursor there first, and a landing alone is nothing more; one
		// on a line not ed's goes with it, for it was meant for that
		// line and not the cursor's
		if ev.Node != "" {
			if !e.reach(ev.Node, ev.At) {
				e.msg = "gone: " + ev.Node
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
		// Ctrl-C is a raw byte here: quit through the guard
		if k.Name == tty.KInt {
			e.quit()
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
