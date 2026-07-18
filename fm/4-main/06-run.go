func (m *fm) run() error {
	if err := os.Chdir(m.cwd); err != nil {
		return err
	}
	t, err := tty.New()
	if err != nil {
		return err
	}
	m.t = t
	ignoreInterrupt()
	t.Write(tty.AltOn + tty.Civis)
	t.Flush()
	m.title("fm:" + m.cwd)
	defer func() {
		t.Write(tty.AltOff + tty.Cnorm)
		t.Flush()
		t.Close()
	}()
	for !m.done {
		m.ls()
		m.draw()
		k, ok := t.ReadKey()
		if !ok || k.Name == tty.KInt {
			return nil
		}
		if k.Name == tty.KResize { // redraw at the top of the loop
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
