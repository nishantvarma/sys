// panelVerb hides the frame's panel, its terminals (layout), their room
// going to the slots above, or shows it again across the frame's bottom:
// bare, it turns it, and hide and show say which. A terminal's tab shown,
// l new's among them, shows the panel with it (show). The frame is the
// pane's, the one it runs in or the one -t names, as alt-t's binding
// names it (`tmux/conf:/alt-t hides/`), which is answered on tmux's
// message line (answered); with no pane it is gio's window's, which keeps
// its panel in a file (askPanel), as tab asks it of its tabs (turn).
func panelVerb(args []string) error {
	pane, named := os.Getenv("TMUX_PANE"), false
	if len(args) > 1 && args[0] == "-t" {
		pane, args, named = args[1], args[2:], true
	}
	if len(args) > 1 || len(args) == 1 && args[0] != "hide" &&
		args[0] != "show" {
		return errors.New("usage: studio panel [-t pane]" +
			" [hide | show]")
	}
	switch {
	case pane == "":
		return askPanel(args)
	case named:
		return answered(pane, turnPanel(pane, args))
	}
	return turnPanel(pane, args)
}
