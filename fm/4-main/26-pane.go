// pane is the window cur is kept per, and whether this pane is the one being
// looked at — one ask, at startup, the eye kept up to date by KFocus after
// that. $TMUX_PANE names a pane and not its window, and an untargeted tmux
// answers about the session's current window and not ours. Outside tmux the
// window is none and the eye is here.
func pane() (string, bool) {
	p := os.Getenv("TMUX_PANE")
	if p == "" {
		return "none", true
	}
	out, err := exec.Command(
		"tmux", "display", "-t", p, "-p",
		"#{window_id} #{pane_active}",
	).Output()
	if err != nil {
		return "none", true
	}
	f := strings.Fields(string(out))
	if len(f) != 2 {
		return "none", true
	}
	return f[0], f[1] == "1"
}
