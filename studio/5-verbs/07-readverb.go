// readVerb prints a slot's screen as text, wrapped lines joined: what a
// program reads back from the shell, or from any slot, by name.
func readVerb(args []string) error {
	if len(args) != 1 {
		return errors.New("usage: studio read slot")
	}
	p, err := paneOf(args[0])
	if err != nil {
		return err
	}
	out, err := (&tmux{}).run("capture-pane", "-p", "-J", "-t", p)
	if err != nil {
		return err
	}
	fmt.Println(out)
	return nil
}
