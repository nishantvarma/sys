// askPanel hides or shows the panel outside tmux, gio's window's, by the
// window's panel file, which the window keeps and follows as it follows
// cur: hidden, or empty for shown; bare, it turns it.
func askPanel(args []string) error {
	win, _ := state.Pane()
	f := state.Path("panel", win)
	hide := state.Read(f) != "hidden"
	if len(args) == 1 {
		hide = args[0] == "hide"
	}
	said := ""
	if hide {
		said = "hidden"
	}
	return state.Write(f, said)
}
