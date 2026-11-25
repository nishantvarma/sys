// slotPane is the pane window win keeps for a slot by its name (open),
// while that pane is there: a pane id, not what else studio keeps under
// the same prefix, the frame's note among it.
func slotPane(t *tmux, win, name string) (string, error) {
	p, err := t.run("show", "-wqv", "-t", win, "@studio-"+name)
	if err != nil {
		return "", err
	}
	if !strings.HasPrefix(p, "%") {
		return "", fmt.Errorf("no slot %s in window %s", name, win)
	}
	if _, err := t.run("display", "-p", "-t", p,
		"#{pane_id}"); err != nil {
		return "", fmt.Errorf("slot %s's pane %s is gone", name, p)
	}
	return p, nil
}
