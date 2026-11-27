// park starts a tab out of sight, in a window of its own in the frame's
// back session (back), made with its first tab, in dir, with env, which
// tells a frame's program its slot and its frame's window, so that it
// keeps the frame's state while it waits (state.Pane). It carries its
// slot's name, its own and its frame's, as a tab shown does (open), and
// stays, dead, when what runs in it ends, for the frame to take it away
// (closeVerb). A session's -e is the session's own, which every window
// made in it after would be told too, a terminal the first tab's slot:
// the back session takes it back once its first tab has it.
func (t *tmux) park(win, dir, slot, name, cmd string,
	env []string) (string, error) {
	b := back(win)
	args := []string{"new-window", "-d", "-t", "=" + b + ":"}
	if _, err := t.run("has-session", "-t", "="+b); err != nil ||
		t.dry && !t.backed {
		args = []string{"new-session", "-d", "-s", b}
	}
	t.backed = true
	made := args[0] == "new-session"
	args = append(args, "-P", "-F", "#{pane_id}", "-c", dir)
	for _, e := range env {
		args = append(args, "-e", e)
	}
	if cmd != "" {
		args = append(args, cmd)
	}
	p, err := t.run(args...)
	if err != nil {
		return "", err
	}
	if made {
		for _, e := range env {
			v, _, _ := strings.Cut(e, "=")
			if _, err := t.run("set-environment", "-t", "="+b,
				"-r", v); err != nil {
				return "", err
			}
		}
	}
	for _, o := range [][2]string{{"@slot", slot}, {"@tab", name},
		{"@frame", win}, {"remain-on-exit", "on"}} {
		if _, err := t.run("set", "-p", "-t", p, o[0],
			o[1]); err != nil {
			return "", err
		}
	}
	return p, nil
}
