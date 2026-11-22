func (m *fm) title(s string) {
	if m.t == nil { // a window titles itself
		return
	}
	m.t.Write(fmt.Sprintf("\x1b]2;%s\x07", s))
	m.t.Flush()
}
