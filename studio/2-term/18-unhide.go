// unhide brings a tab of the hidden panel back across the frame's bottom,
// 38% of its height, as the layout first gave it (layout).
func (t *tmux) unhide(win, pane string) error {
	first, err := t.run("display", "-p", "-t", win, "#{pane_id}")
	if err != nil {
		return err
	}
	pct := strconv.Itoa(100-golden) + "%"
	if _, err := t.run("join-pane", "-d", "-f", "-v", "-l", pct, "-s",
		pane, "-t", first); err != nil {
		return err
	}
	_, err = t.run("set", "-wu", "-t", win, "@studio-hidden")
	return err
}
