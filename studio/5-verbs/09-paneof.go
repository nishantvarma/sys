// paneOf is the pane the studio window keeps for a slot, by its name.
func paneOf(name string) (string, error) {
	win, err := target()
	if err != nil {
		return "", err
	}
	p, err := (&tmux{}).run("show", "-wqv", "-t", win, "@studio-"+name)
	switch {
	case err != nil:
		return "", err
	case p == "":
		return "", fmt.Errorf("no slot %s in window %s", name, win)
	}
	return p, nil
}
