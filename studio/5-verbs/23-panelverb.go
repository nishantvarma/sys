// panelVerb hides the frame's panel, its terminals (layout), their room
// going to the slots above, or shows it again across the frame's bottom:
// bare, it turns it, and hide and show say which. A terminal's tab shown,
// l new's among them, shows the panel with it (show). The frame is the
// pane's, the one it runs in or the one -t names, as alt-t's binding
// names it (`tmux/conf:/alt-t hides/`); with no pane it is gio's
// window's, which keeps its panel in a file (askPanel), as tab asks it
// of its tabs (turn).
func panelVerb(args []string) error {
	pane := os.Getenv("TMUX_PANE")
	if len(args) > 1 && args[0] == "-t" {
		pane, args = args[1], args[2:]
	}
	if len(args) > 1 || len(args) == 1 && args[0] != "hide" &&
		args[0] != "show" {
		return errors.New("usage: studio panel [-t pane]" +
			" [hide | show]")
	}
	if pane == "" {
		return askPanel(args)
	}
	t := &tmux{}
	win, _ := t.run("display", "-p", "-t", pane, "#{@frame}")
	if win == "" {
		w, err := target()
		if err != nil {
			return err
		}
		win = w
	}
	slot, _ := t.run("show", "-wqv", "-t", win, "@studio-panel")
	if slot == "" {
		return errors.New("the frame has no panel")
	}
	tps, err := t.tabsOf(win, slot)
	if err != nil {
		return err
	}
	at := slices.IndexFunc(tps, func(tp tabPane) bool {
		return tp.window == win
	})
	hide := at >= 0
	if len(args) == 1 {
		hide = args[0] == "hide"
	}
	switch {
	case hide && at >= 0:
		return t.hide(win, tps[at].pane)
	case !hide && at < 0 && len(tps) > 0:
		was, _ := t.run("show", "-wqv", "-t", win, "@studio-hidden")
		tp := tps[0]
		for _, o := range tps {
			if o.pane == was {
				tp = o
			}
		}
		return t.show(win, slot, tp)
	}
	return nil
}
