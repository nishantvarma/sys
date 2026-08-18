func (m *fm) title(s string) {
	m.t.Write(fmt.Sprintf("\x1b]2;%s\x07", s))
	m.t.Flush()
}
