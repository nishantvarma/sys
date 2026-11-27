// closeTab ends the tab of a slot shown at at, of its tabs tps: the next
// live one takes its place first (next), as when it ends itself (died);
// with none, it goes, and the slot's room with it. The tab leaves its slot,
// and the slot's bar names it no more, before it is killed, for the kill
// ends studio too when studio runs in the tab's own shell, as l close
// does there.
func (t *tmux) closeTab(win, slot string, tps []tabPane, at int) error {
	shown := tps[at]
	if n, ok := next(tps, at); ok {
		if err := t.show(win, slot, n); err != nil {
			return err
		}
	}
	if _, err := t.run("set", "-pu", "-t", shown.pane,
		"@slot"); err != nil {
		return err
	}
	if err := t.bars(win, slot); err != nil {
		return err
	}
	_, err := t.run("kill-pane", "-t", shown.pane)
	return err
}
