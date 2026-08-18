func (m *fm) run() error {
	if err := os.Chdir(m.cwd); err != nil {
		return err
	}
	t, err := tty.New()
	if err != nil {
		return err
	}
	m.t = t
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
	for !m.done {
		m.ls()
		m.say()
		m.draw()
		k, ok := t.ReadKey()
		if !ok || k.Name == tty.KInt {
			return nil
		}
		if k.Name == tty.KResize { // redraw at the top of the loop
			continue
		}
		if k.Name == tty.KFocus || k.Name == tty.KBlur {
			m.on = k.Name == tty.KFocus
			// a still row is silent, the eye coming back is
			// not: another pane spoke while we were away.
			m.said = ""
			continue
		}
		m.msg = ""
		if b, hit := m.keys[k.Token()]; hit {
			b.act()
		} else if k.Rune >= '0' && k.Rune <= '9' {
			m.jump(k.Rune)
		}
	}
	return nil
}
