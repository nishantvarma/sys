// hide parks the panel's shown pane out of sight, as a tab waits, in the
// frame's back session (backed), its room going to the slots above, and
// the window keeps it, @studio-hidden, for the panel to show it again
// where it was (hidden).
func (t *tmux) hide(win, pane string) error {
	wait, err := t.backed(win)
	if err != nil {
		return err
	}
	for _, c := range [][]string{
		{"break-pane", "-d", "-s", pane, "-t", "=" + back(win) + ":"},
		{"set", "-w", "-t", win, "@studio-hidden", pane},
	} {
		if _, err := t.run(c...); err != nil {
			return err
		}
	}
	if wait != "" {
		t.run("kill-pane", "-t", wait)
	}
	return nil
}
