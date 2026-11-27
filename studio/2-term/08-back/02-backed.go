// backed makes the frame's back session (back), for a tab to wait in,
// when there is none: with a window of its own that waits for the tab,
// whose pane comes back for the caller to kill once the tab is in, for a
// session lives only while it has a window; none when the session was
// there, or another verb made it at once.
func (t *tmux) backed(win string) (string, error) {
	b := back(win)
	if _, err := t.run("has-session", "-t", "="+b); err == nil &&
		(!t.dry || t.made) {
		return "", nil
	}
	t.made = true
	wait, err := t.run("new-session", "-d", "-s", b, "-P", "-F",
		"#{pane_id}")
	if err != nil {
		if _, e := t.run("has-session", "-t", "="+b); e == nil {
			return "", nil
		}
		return "", err
	}
	return wait, nil
}
