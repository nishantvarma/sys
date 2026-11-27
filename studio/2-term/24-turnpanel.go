// turnPanel hides or shows the panel of the frame a pane is in, in tmux,
// as panelVerb asks: hidden, it shows the tab it hid (hidden).
func turnPanel(pane string, args []string) error {
	t := &tmux{}
	win, _ := t.run("display", "-p", "-t", pane, "#{@frame}")
	if win == "" {
		return errors.New("not in a frame")
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
		return t.show(win, slot, t.hidden(win, tps))
	}
	return nil
}
