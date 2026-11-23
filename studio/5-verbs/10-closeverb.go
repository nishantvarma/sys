// closeVerb ends a studio window as its slots end themselves: fm, ed and
// a lens are each asked once, by the keys they quit on, Esc to end what
// they were in the middle of and then q, so ed keeps an unsaved buffer,
// saying so, until a save or a second q; a shell waiting at its prompt
// ends, and one running a command is left to finish. A slot's end runs
// it, so one q ends the frame (open). fm and ed are as they were.
func closeVerb(args []string) error {
	if len(args) > 1 {
		return errors.New("usage: studio close [window]")
	}
	win, err := "", error(nil)
	if len(args) == 1 {
		win = args[0]
	} else if win, err = target(); err != nil {
		return err
	}
	t := &tmux{}
	out, err := t.run("list-panes", "-t", win, "-F", "#{pane_id} "+
		"#{pane_dead} #{?@studio-asked,1,0} #{pane_current_command}")
	if err != nil {
		return nil // the window has ended already
	}
	sh, _ := t.run("show", "-gv", "default-shell")
	for _, l := range strings.Split(out, "\n") {
		f := strings.Fields(l)
		if len(f) != 4 || f[1] == "1" {
			continue // a pane gone
		}
		pane, asked, cmd := f[0], f[2] == "1", f[3]
		switch {
		case cmd == filepath.Base(sh):
			t.run("kill-pane", "-t", pane)
		case !asked && (cmd == "fm" || cmd == "ed" || cmd == "studio"):
			t.run("set", "-p", "-t", pane, "@studio-asked", "1")
			t.run("send-keys", "-t", pane, "Escape", "q")
		}
	}
	return nil
}
