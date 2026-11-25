// openVerb lays a frame out in the folder studio runs in: studio's own
// frame, or the note given. Run from a shell alone in its tmux window, it
// lays the frame out there, around that shell, when the frame has one;
// else in a new window. -n prints the tmux commands and runs none.
func openVerb(args []string) error {
	dry := len(args) > 0 && args[0] == "-n"
	if dry {
		args = args[1:]
	}
	if len(args) > 1 {
		return errors.New("usage: studio [-n] [frame]")
	}
	f, err := load(args)
	if err != nil {
		return err
	}
	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	here := ""
	if p := os.Getenv("TMUX_PANE"); p != "" && f.shell() != nil {
		n, err := (&tmux{}).run("display", "-p", "-t", p,
			"#{window_panes}")
		if err == nil && n == "1" {
			here = p
		}
	}
	return f.open(dir, here, dry)
}
