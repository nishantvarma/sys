// ask is what a verb asks of a frame outside tmux, added to the window's
// ask file for the frame's host, gio's window, to do in order (asked). A
// program in a slot is told it, $STUDIO (state.Slot); one told none, a
// terminal's shell, asks of the slot it is typed in, the one with the
// eye, ".", as tmux finds a shell's by its pane. What gio starts is told
// its frame's window, $STUDIO_WINDOW, so that a program told neither is
// in no frame, and asks nothing of a window it is not in.
func ask(verb, slot string, words []string) error {
	if slot == "" && os.Getenv("STUDIO_WINDOW") == "" {
		return errors.New("not in a frame")
	}
	if slot == "" {
		slot = "."
	}
	line, err := json.Marshal(asked{verb, slot, words})
	if err != nil {
		return err
	}
	win, _ := state.Pane()
	return state.Add(state.Path("ask", win), string(line))
}
