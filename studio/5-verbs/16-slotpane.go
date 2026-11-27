// slotPane is the pane of window win that carries a slot's name, @slot
// (open), the tab the slot shows; else the pane of a tab by that name,
// @tab, wherever in the frame it waits. tmux lists only the panes that are
// there, so the name of a pane that ended went with it.
func slotPane(t *tmux, win, name string) (string, error) {
	for _, find := range [][]string{
		{"-t", win, "-f", "#{==:#{@slot}," + name + "}"},
		{"-a", "-f", "#{&&:#{==:#{@frame}," + win + "},#{==:#{@tab}," +
			name + "}}"},
	} {
		out, err := t.run(append(append([]string{"list-panes"},
			find...), "-F", "#{pane_id}")...)
		if err != nil {
			return "", err
		}
		if p, _, _ := strings.Cut(out, "\n"); p != "" {
			return p, nil
		}
	}
	return "", fmt.Errorf("no slot or tab %s in window %s", name, win)
}
