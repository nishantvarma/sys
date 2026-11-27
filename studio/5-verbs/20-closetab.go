// closeTab ends the tab of a slot shown at at, of its tabs tps: the next
// live one takes its place first, as when it ends itself (died); with
// none, it goes, and the slot's room with it.
func (t *tmux) closeTab(win, slot string, tps []tabPane, at int) error {
	shown := tps[at]
	for i := 1; i < len(tps); i++ {
		if n := tps[(at+i)%len(tps)]; !n.dead {
			if err := t.show(win, slot, n); err != nil {
				return err
			}
			break
		}
	}
	if _, err := t.run("kill-pane", "-t", shown.pane); err != nil {
		return err
	}
	return t.bars(win, slot)
}
