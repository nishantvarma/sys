// hide parks the panel's shown pane out of sight, as a tab waits (park),
// its room going to the slots above, and the window keeps it, @studio-hidden,
// for the panel to show it again where it was (panelVerb).
func (t *tmux) hide(win, pane string) error {
	b := back(win)
	if _, err := t.run("has-session", "-t", "="+b); err != nil {
		if _, err := t.run("new-session", "-d", "-s", b); err != nil {
			return err
		}
	}
	for _, c := range [][]string{
		{"break-pane", "-d", "-s", pane, "-t", "=" + b + ":"},
		{"set", "-w", "-t", win, "@studio-hidden", pane},
	} {
		if _, err := t.run(c...); err != nil {
			return err
		}
	}
	return nil
}
