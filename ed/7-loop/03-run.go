// run is ed on the terminal: a screen of its own, a paste as one key, the
// mouse its own, and the shell's screen back when it is done.
func (e *editor) run() error {
	t, err := tty.New()
	if err != nil {
		return err
	}
	e.t = t
	e.ui = &term{t: t}
	e.wake()
	t.Write(tty.AltOn + tty.PasteOn)
	t.Mouse()
	defer func() {
		t.Write(tty.MouseOff + tty.PasteOff + tty.AltOff)
		t.Write(tty.Cnorm + tty.CurReset)
		t.Flush()
		t.Close()
	}()
	e.loop()
	return nil
}
