// sendVerb pastes text into a slot and presses enter, as p -a pastes into
// the shell (`sh/p`): a program writes to the shell, or to any slot, by
// name. The paste is bracketed for a program that asks, so a shell takes
// a newline in it as text, and enter is what runs it. Text that starts
// with - goes as it is.
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
	for _, c := range [][]string{
		{"set-buffer", "-b", "studio", "--", text},
		{"paste-buffer", "-d", "-p", "-b", "studio", "-t", p},
		{"send-keys", "-t", p, "Enter"},
	} {
		if _, err := t.run(c...); err != nil {
			return err
		}
	}
	return nil
}
