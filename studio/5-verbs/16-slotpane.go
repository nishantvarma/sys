// slotPane is the pane of window win that carries a slot's name, @slot
// (open). tmux lists only the panes that are there, so the name of a pane
// that ended went with it.
func slotPane(t *tmux, win, name string) (string, error) {
	out, err := t.run("list-panes", "-t", win, "-f",
		"#{==:#{@slot},"+name+"}", "-F", "#{pane_id}")
	if err != nil {
		return "", err
	}
	if p, _, _ := strings.Cut(out, "\n"); p != "" {
		return p, nil
	}
	return "", fmt.Errorf("no slot %s in window %s", name, win)
}
