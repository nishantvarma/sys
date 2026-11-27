// turnPanel hides or shows the panel of the frame a pane is in, in tmux,
// as panelVerb asks: else the studio a program in another window reaches
// (target).
func turnPanel(pane string, args []string) error {
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
