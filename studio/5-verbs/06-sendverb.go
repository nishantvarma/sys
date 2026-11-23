// sendVerb types text into a slot and presses enter, as p -a types into
// the marked pane: a program writes to the shell, or to any slot, by
// name. Text that starts with - is typed as it is.
func sendVerb(args []string) error {
	if len(args) < 2 {
		return errors.New("usage: studio send slot text...")
	}
	p, err := paneOf(args[0])
	if err != nil {
		return err
	}
	t := &tmux{}
	text := strings.Join(args[1:], " ")
	if _, err := t.run("send-keys", "-t", p, "-l", "--",
		text); err != nil {
		return err
	}
	_, err = t.run("send-keys", "-t", p, "Enter")
	return err
}
