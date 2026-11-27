// tabsOf is a slot's tabs, wherever each waits, in the order they were
// made, as their panes' ids count: the frame's own first, and a terminal
// opened later after them. A tab that ended out of sight goes as it is
// found, for nothing is left to show of it.
func (t *tmux) tabsOf(win, slot string) ([]tabPane, error) {
	out, err := t.run("list-panes", "-a", "-f",
		"#{&&:#{==:#{@frame},"+win+"},#{==:#{@slot},"+slot+"}}",
		"-F", "#{pane_id} #{window_id} #{pane_dead} #{@tab}"+
			" #{pane_current_command}")
	if err != nil {
		return nil, err
	}
	var tps []tabPane
	for _, l := range strings.Split(out, "\n") {
		f := strings.Fields(l)
		if len(f) < 4 {
			continue
		}
		tp := tabPane{f[0], f[1], f[3], f[2] == "1",
			strings.Join(f[4:], " ")}
		if tp.dead && tp.window != win {
			t.run("kill-pane", "-t", tp.pane)
			continue
		}
		tps = append(tps, tp)
	}
	slices.SortFunc(tps, func(a, b tabPane) int {
		return paneNumber(a.pane) - paneNumber(b.pane)
	})
	return tps, nil
}
