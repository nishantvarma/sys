// newTab opens a terminal in the frame's panel, as l new asks (turn):
// named by the next number free among the panel's tabs, running cmd, a
// shell for none, in the folder of the pane it is asked from, as p takes
// the folder that printed its word (`sh/p:/the pane's own cwd/`), and
// told nothing of the frame, for a terminal's programs are in none
// (told). It shows it, and the panel with it when it was hidden (show);
// it stays, dead, as what runs in it ends, for the next tab to take its
// place (closeVerb), as the panel's first does (open). A frame laid out
// before it kept tabs names no panel, and says so, and spawn splits a
// pane there as it would anywhere.
func (t *tmux) newTab(win, pane string, cmd []string) error {
	panel, _ := t.run("show", "-wqv", "-t", win, "@studio-panel")
	if panel == "" {
		return errors.New("the frame has no panel: lay it out again")
	}
	tps, err := t.tabsOf(win, panel)
	if err != nil {
		return err
	}
	taken := map[string]bool{}
	for _, tp := range tps {
		taken[tp.name] = true
	}
	n := len(tps) + 1
	for taken[strconv.Itoa(n)] {
		n++
	}
	dir, err := t.run("display", "-p", "-t", pane,
		"#{pane_current_path}")
	if err != nil {
		return err
	}
	p, err := t.park(win, dir, panel, strconv.Itoa(n), cmd, nil)
	if err != nil {
		return err
	}
	return t.show(win, panel, tabPane{pane: p, name: strconv.Itoa(n)})
}
