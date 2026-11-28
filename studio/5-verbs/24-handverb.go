// handVerb hands text to the shell of gio's window, at its prompt, as run
// does in tmux itself (`sh/run`): the window keeps its terminals, and a
// program in its frame has no pane to paste into, so the window is asked
// (ask), as panel asks of its panel, and pastes it as a middle click
// there does, where only Return runs it.
func handVerb(args []string) error {
	if len(args) == 0 {
		return errors.New("usage: studio hand text...")
	}
	return ask("hand", "", []string{strings.Join(args, " ")})
}
