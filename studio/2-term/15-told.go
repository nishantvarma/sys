// told is what a frame's program is told as its pane is made: its slot,
// $STUDIO (state.Slot), and its frame's window, $STUDIO_WINDOW, for one
// that starts out of sight (state.Pane). A shell is told nothing, for
// what runs from it is in no frame.
func told(slot, win, cmd string) []string {
	if cmd == "" {
		return nil
	}
	return []string{"STUDIO=" + slot, "STUDIO_WINDOW=" + win}
}
