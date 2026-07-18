// help shows the keymap via tty.Help, built from normalKeys — the same
// table dispatch reads, so the screen and the bindings can't drift.
func (e *editor) help() {
	var keys [][2]string
	for _, b := range normalKeys() {
		keys = append(keys, [2]string{b.key, b.desc})
	}
	tty.Help(e.t, "ed — keys", keys)
}
