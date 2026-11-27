// menu opens tmux's own menu of a slot's tabs, tps, the one at at shown,
// at the slot's left, under the names it lists: each tab by its place, the
// key that picks it, and its name, or, a terminal's, what runs in it; then,
// for the panel, a new terminal, the shown one closed, and the panel
// hidden. A pick runs studio tab, as l does. Given a client, the menu is
// that client's, as a press on the bar names it. tmux gives a menu the
// mouse only when a click's binding opens it, and studio opens this one,
// so it asks for the mouse, -M, and to stay open as the click's release
// comes, -O: a press on an item picks it, and one off the menu closes it.
func (t *tmux) menu(win, slot, client string, tps []tabPane,
	at int) error {
	if at < 0 || len(tps) == 0 {
		return nil
	}
	self, err := os.Executable()
	if err != nil {
		return err
	}
	shown := tps[at].pane
	run := func(words ...string) string {
		cmd := quote(append([]string{self}, words...))
		return "run-shell -b " + quote([]string{cmd})
	}
	panel, _ := t.run("show", "-wqv", "-t", win, "@studio-panel")
	args := []string{"display-menu", "-M", "-O", "-t", shown, "-x",
		"#{popup_pane_left}", "-y", "#{popup_pane_top}", "-T",
		" " + slot + " "}
	if client != "" {
		args = append(args, "-c", client)
	}
	for i, tp := range tps {
		n, key, label := strconv.Itoa(i+1), "", tp.name
		if i < 9 {
			key = n
		}
		if slot == panel && tp.cmd != "" {
			label = tp.cmd
		}
		args = append(args, n+"  "+label, key,
			run("tab", "-t", shown, n))
	}
	if slot == panel {
		args = append(args, "",
			"+  new terminal", "n", run("tab", "-t", shown, "new"),
			"×  close", "x", run("tab", "-t", shown, "close"),
			"–  hide the panel", "h",
			run("panel", "-t", shown, "hide"))
	}
	_, err = t.run(args...)
	return err
}
