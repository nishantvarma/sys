// target is the studio window a verb is for: the caller's own, when
// studio laid it out, else the one studio laid out last, while it is
// there, so that a program in any window, an agent's say, reaches it.
func target() (string, error) {
	t := &tmux{}
	if w, _ := state.Pane(); w != "none" {
		f, err := t.run("show", "-wqv", "-t", w, "@studio-frame")
		if err != nil {
			return "", err
		}
		if f != "" {
			return w, nil
		}
	}
	w, err := t.run("show", "-gqv", "@studio")
	switch {
	case err != nil:
		return "", err
	case w == "":
		return "", errors.New("no studio window yet; run studio")
	}
	// tmux answers a window gone with an error, or with nothing
	if id, err := t.run("display", "-p", "-t", w,
		"#{window_id}"); err != nil || id != w {
		return "", fmt.Errorf("studio window %s is gone; run studio",
			w)
	}
	return w, nil
}
