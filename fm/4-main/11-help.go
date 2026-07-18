// help shows fm's shortcuts via tty.Help, built from the bind order.
func (m *fm) help() {
	var keys [][2]string
	for _, b := range m.order {
		keys = append(keys, [2]string{b.key, b.desc})
	}
	tty.Help(m.t, "Shortcuts", keys)
}
