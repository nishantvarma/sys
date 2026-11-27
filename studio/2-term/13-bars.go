// bars puts a slot's tabs on the border over each of its panes, @bar, the
// one in the frame's window lit; and, once a slot has tabs to name, the
// frame's window draws each pane's border so, over it, which says nothing
// over a pane with no tabs to show. A frame of no tabs keeps its rows.
func (t *tmux) bars(win, slot string) error {
	tps, err := t.tabsOf(win, slot)
	if err != nil {
		return err
	}
	panel, _ := t.run("show", "-wqv", "-t", win, "@studio-panel")
	var names []string
	shown := 0
	for i, tp := range tps {
		name := tp.name
		if slot == panel {
			name = strconv.Itoa(i + 1) // a terminal, by its place
		}
		names = append(names, name)
		if tp.window == win {
			shown = i
		}
	}
	b := bar(names, shown)
	for _, tp := range tps {
		if _, err := t.run("set", "-p", "-t", tp.pane, "@bar",
			b); err != nil {
			return err
		}
	}
	if b == "" {
		return nil
	}
	for _, o := range [][2]string{{"pane-border-status", "top"},
		{"pane-border-format", "#{@bar}"}} {
		if _, err := t.run("set", "-w", "-t", win, o[0],
			o[1]); err != nil {
			return err
		}
	}
	return nil
}
