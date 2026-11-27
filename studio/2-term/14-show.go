// show brings a tab of a slot into the slot's place in the frame's window:
// the pane shown there trades places with the tab's, which parks it out of
// sight (park), and the eye goes to the tab, so that keys reach it; with
// none shown, the slot is the panel, hidden, and the tab shows it again
// (unhide). Then the slot's bar lights it (bars).
func (t *tmux) show(win, slot string, tp tabPane) error {
	tps, err := t.tabsOf(win, slot)
	if err != nil {
		return err
	}
	shown := false
	for _, o := range tps {
		if o.window != win {
			continue
		}
		shown = true
		if o.pane != tp.pane {
			if _, err := t.run("swap-pane", "-d", "-s", tp.pane,
				"-t", o.pane); err != nil {
				return err
			}
		}
	}
	if !shown {
		if err := t.unhide(win, tp.pane); err != nil {
			return err
		}
	}
	if err := t.bars(win, slot); err != nil {
		return err
	}
	_, err = t.run("select-pane", "-t", tp.pane)
	return err
}
