// ask is what is asked of a slot's tabs outside tmux, put in the window's
// tab file for the frame's host, gio's window, to do, as it follows cur:
// the slot, the ask, and when, so that one asked twice is done twice. A
// program told no slot, a terminal's shell, asks of the slot it is typed
// in, the one with the eye, ".", as tmux finds a shell's by its pane;
// and the window's tabs are said on its bar, not here.
func ask(slot, word string) error {
	if slot == "" {
		slot = "."
	}
	if word == "" {
		return errors.New("the window says its tabs on its bar")
	}
	win, _ := state.Pane()
	line := fmt.Sprintf("%s %s %d", slot, word, time.Now().UnixNano())
	return state.Write(state.Path("tab", win), line)
}
