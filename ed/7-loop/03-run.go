func (e *editor) run() error {
	t, err := tty.New()
	if err != nil {
		return err
	}
	e.t = t
	t.Write(tty.AltOn + tty.PasteOn)
	t.Flush()
	defer func() {
		t.Write(tty.PasteOff + tty.AltOff)
		t.Write(tty.Cnorm + tty.CurReset)
		t.Flush()
		t.Close()
	}()
	for !e.done {
		// Only the last frame of a burst is worth painting, and a
		// paste is a long burst: draw once the queue has run dry.
		if !t.Pending() {
			e.draw()
		}
		k, ok := t.ReadKey()
		if !ok {
			break
		}
		if k.Name == tty.KResize { // redraw at the top of the loop
			continue
		}
		// Ctrl-C is a raw byte here: quit through the guard
		if k.Name == tty.KInt {
			e.quit()
			continue
		}
		e.msg = ""
		// A paste is text, not keystrokes: it goes in whole, in
		// either mode, and never through a binding.
		if k.Name == tty.KPaste {
			e.pasteText(k.Text)
			continue
		}
		e.mode.key(e, k)
	}
	return nil
}
