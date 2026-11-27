// died is a tab that ended in its slot's place, and stayed there, dead,
// for this: the next live tab of the slot takes its place, which parks it
// out of sight, where it goes as the slot's tabs are next read (tabsOf);
// with none, it goes here, and the slot's room with it, as a pane's that
// ends.
func (t *tmux) died(win, pane string) error {
	slot, _ := t.run("display", "-p", "-t", pane, "#{@slot}")
	tps, err := t.tabsOf(win, slot)
	if err != nil {
		return err
	}
	for _, tp := range tps {
		if tp.pane != pane && !tp.dead {
			return t.show(win, slot, tp)
		}
	}
	_, err = t.run("kill-pane", "-t", pane)
	return err
}
