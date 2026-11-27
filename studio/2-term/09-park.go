// park starts a tab out of sight, in a window of its own in the frame's
// back session (backed), in dir, running cmd, a shell for none, with env,
// which tells a frame's program its slot and its frame's window, so that
// it keeps the frame's state while it waits (state.Pane). It carries its
// slot's name, its own and its frame's, as a tab shown does (open), and
// stays, dead, when what runs in it ends, for the frame to take it away
// (closeVerb).
func (t *tmux) park(win, dir, slot, name string, cmd,
	env []string) (string, error) {
	wait, err := t.backed(win)
	if err != nil {
		return "", err
	}
	args := []string{"new-window", "-d", "-t", "=" + back(win) + ":",
		"-P", "-F", "#{pane_id}", "-c", dir}
	for _, e := range env {
		args = append(args, "-e", e)
	}
	p, err := t.run(append(args, cmd...)...)
	if err != nil {
		return "", err
	}
	if wait != "" {
		t.run("kill-pane", "-t", wait)
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
