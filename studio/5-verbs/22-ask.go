// ask is what is asked of a slot's tabs outside tmux, put in the window's
// tab file for the frame's host, gio's window, to do, as it follows cur:
// the slot, the ask, and when, so that one asked twice is done twice. A
// terminal is tmux's, and the window has none; and its tabs are said on
// its bar, where the window keeps them, not here.
func ask(slot, word string) error {
	switch {
	case slot == "":
		return errors.New("not in a frame's slot")
	case word == "new":
		return errors.New("tab new: a terminal is tmux's, and the" +
			" window has none")
	case word == "":
		return errors.New("the window says its tabs on its bar")
	}
	win, _ := state.Pane()
	line := fmt.Sprintf("%s %s %d", slot, word, time.Now().UnixNano())
	return state.Write(state.Path("tab", win), line)
}
