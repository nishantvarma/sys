// jump selects a file by its displayed number, auto-committing once no longer
// number could stay in range.
func (m *fm) jump(first rune) {
	_, h := m.t.Size()
	buf := string(first)
	n := len(m.files)
	m.t.Write(tty.Status(h, buf, true))
	m.t.Flush()
	for {
		if v, _ := strconv.Atoi(buf); v*10 > n {
			break
		}
		k, ok := m.t.ReadKey()
		if !ok || k.Name == tty.KEsc || k.Name == tty.KInt {
			m.t.Write(tty.Civis)
			m.t.Flush()
			return
		}
		if k.Name == tty.KEnter || k.Rune < '0' || k.Rune > '9' {
			break
		}
		buf += string(k.Rune)
		m.t.Write(string(k.Rune))
		m.t.Flush()
	}
	m.t.Write(tty.Civis)
	m.t.Flush()
	if v, err := strconv.Atoi(buf); err == nil && v >= 1 && v <= n {
		m.idx = v - 1
	}
}
