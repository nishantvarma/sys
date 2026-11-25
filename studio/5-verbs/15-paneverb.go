// paneVerb prints the pane that holds a slot of the caller's own frame,
// by the slot's name: how e, o and p -a reach the frame's ed, fm and
// shell in place of a pane beside (`design/mouse`). Only the caller's
// window, for cur, which the slots follow, is the window's. Outside a
// frame, or with no slot by that name, it fails, and the caller opens a
// pane of its own.
func paneVerb(args []string) error {
	if len(args) != 1 {
		return errors.New("usage: studio pane slot")
	}
	win, _ := state.Pane()
	if win == "none" {
		return errors.New("studio pane: not in tmux")
	}
	p, err := slotPane(&tmux{}, win, args[0])
	if err != nil {
		return err
	}
	fmt.Println(p)
	return nil
}
