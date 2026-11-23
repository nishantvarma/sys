// openVerb lays a frame out in a new tmux window, in the folder studio
// runs in: studio's own frame, or the note given; -n prints the tmux
// commands and runs none.
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
	return f.open(dir, dry)
}
