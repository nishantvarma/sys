func (e *editor) run() error {
	t, err := tty.New()
	if err != nil {
		return err
	}
	e.t = t
	t.Write(tty.AltOn)
	t.Flush()
	defer func() {
		t.Write(tty.AltOff + tty.Cnorm + tty.CurReset)
		t.Flush()
		t.Close()
	}()
	for !e.done {
		e.draw()
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
		e.mode.key(e, k)
	}
	return nil
}
