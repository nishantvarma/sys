// help shows fm's shortcuts, built from the bind order: on the terminal
// by tty.Help, and off it, in a window, as nodes its view lays out
// (outline.Help).
func (m *fm) help() {
	var keys [][2]string
	for _, b := range m.order {
		keys = append(keys, [2]string{b.key, b.desc})
	}
	if m.t == nil {
		outline.Help(m.ui, "Shortcuts", keys)
		return
	}
	tty.Help(m.t, "Shortcuts", keys)
}
