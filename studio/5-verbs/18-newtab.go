// newTab opens a tab in the frame's terminals' slot, the slot that keeps
// its shell (shell): named by the next number free among the slot's tabs,
// running cmd, else a shell, in the folder the verb runs in, and told
// nothing of the frame, for a terminal's programs are in none (told). It
// shows it, and each of the slot's tabs then stays, dead, as what runs in
// it ends, for the next to take its place (closeVerb); and the panel
// shows with it, when it was hidden (show). A frame laid out before it
// kept tabs, which names no panel, says so, and spawn splits a pane there
// as it would anywhere.
func (t *tmux) newTab(win, cmd string) error {
	note, err := t.run("show", "-wqv", "-t", win, "@studio-frame")
	if err != nil {
		return err
	}
	f, err := load([]string{note})
	if err != nil {
		return err
	}
	sh := f.shell()
	if sh == nil {
		return errors.New("the frame has no terminals: no slot of a" +
			" shell")
	}
	tps, err := t.tabsOf(win, sh.name)
	if err != nil {
		return err
	}
	if panel, _ := t.run("show", "-wqv", "-t", win,
		"@studio-panel"); panel == "" {
		return errors.New("the frame keeps no tabs: lay it out again")
	}
	taken := map[string]bool{}
	for _, tp := range tps {
		taken[tp.name] = true
	}
	n := len(tps) + 1
	for taken[strconv.Itoa(n)] {
		n++
	}
	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	p, err := t.park(win, dir, sh.name, strconv.Itoa(n), cmd, nil)
	if err != nil {
		return err
	}
	for _, tp := range tps {
		t.run("set", "-p", "-t", tp.pane, "remain-on-exit", "on")
	}
	return t.show(win, sh.name, tabPane{pane: p, name: strconv.Itoa(n)})
}
