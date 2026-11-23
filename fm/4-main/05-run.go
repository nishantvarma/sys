// run is fm on the terminal: sys/view, on the whole of tty, is its view.
func (m *fm) run() error {
	if err := os.Chdir(m.cwd); err != nil {
		return err
	}
	t, err := tty.New()
	if err != nil {
		return err
	}
	m.t = t
	m.ui = view.New(t, userHome, looks)
	catchInterrupt()
	t.Write(tty.AltOn + tty.Civis)
	t.Flush()
	t.Focus()
	m.title("fm:" + m.cwd)
	defer func() {
		t.Write(tty.FocusOff + tty.AltOff + tty.Cnorm)
		t.Flush()
		t.Close()
	}()
	m.ls()
	m.land(m.pos[m.cwd]) // the argument's own name, else past ..
	m.loop()
	return nil
}
