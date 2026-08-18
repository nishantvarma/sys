// say writes the row to cur, the noun every verb reads when it is given
// none. Only while this pane holds the eye: a background fm — scrolled, or
// woken by a resize — must not take the shell's this. Temp then rename, so a
// reader between the two gets the old address and never an empty one. The
// log is c's alone: a view moves this on every j, and at a line a keypress
// back would be unusable.
func (m *fm) say() {
	c := m.cur()
	if !m.on || c == "" || c == m.said {
		return
	}
	if err := os.MkdirAll(filepath.Dir(m.this), 0o755); err != nil {
		return
	}
	tmp := m.this + "." + strconv.Itoa(os.Getpid())
	if err := os.WriteFile(tmp, []byte(c+"\n"), 0o644); err != nil {
		return
	}
	if err := os.Rename(tmp, m.this); err == nil {
		m.said = c
	}
}
