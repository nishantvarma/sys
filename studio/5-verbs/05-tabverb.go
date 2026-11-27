// tabVerb turns a slot's tabs (`design/lens`) as l asks (turn): the slot
// being the one its caller says it runs in, $STUDIO (state.Slot), else the
// one the pane it runs in names, or the pane -t names, @slot, as a press
// on a frame's tab line does (`tmux/conf:/a frame's tab line/`), which
// is answered on tmux's message line (answered).
func tabVerb(args []string) error {
	if len(args) > 1 && args[0] == "-t" {
		return answered(args[1], turn(state.Slot(), args[1], args[2:]))
	}
	return turn(state.Slot(), os.Getenv("TMUX_PANE"), args)
}
