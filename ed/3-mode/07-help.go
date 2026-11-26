// help shows the keymap, built from normalKeys — the same table dispatch
// reads, so the screen and the bindings can't drift: on the terminal by
// tty.Help, and off it, in a window, as nodes its view lays out
// (outline.Help).
func (e *editor) help() {
	var keys [][2]string
	for _, b := range normalKeys() {
		keys = append(keys, [2]string{b.key, b.desc})
	}
	if e.t == nil {
		outline.Help(e.ui, "ed — keys", keys)
		return
	}
	tty.Help(e.t, "ed — keys", keys)
}
