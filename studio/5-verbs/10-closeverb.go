// closeVerb is what a frame's window runs as a pane of it ends (open). When
// the pane was the frame's face, fm's slot (face), the frame ends with
// it, at once, wherever each of its panes is, in the window or out of
// sight (back): each slot's program is killed, and each shell waiting at
// its prompt. What stays is the shell the frame was laid out around,
// @studio-here, a shell running a command, and a terminal's tab running
// one, which l new and spawn open, an ed among them, which keeps no mark
// of work unsaved (told); one out of sight comes back to the window, which
// is then a plain one again (plain). But while ed names a file it holds
// unsaved, which fm's second q takes away as it lets the work go
// (`fm/4-main/40-quit.go`), the face ended some other way, as a crash
// ends it, and the frame stays for ed to save, saying so. Any other pane's
// end is its own: a shell that exits ends alone, and a tab's, which stays
// dead, gives its place to the next tab of its slot (died).
func closeVerb(args []string) error {
	if len(args) != 2 {
		return errors.New("usage: studio close window pane")
	}
	win, ended := args[0], args[1]
	t := &tmux{}
	if dead, _ := t.run("display", "-p", "-t", ended,
		"#{pane_dead}"); dead == "1" {
		return t.died(win, ended)
	}
	face, err := t.run("show", "-wqv", "-t", win, "@studio-face")
	if err != nil || face != ended {
		return nil // the window is gone, or the pane was not its face
	}
	if f := state.Read(state.Path("unsaved", win)); f != "" {
		say := "fm ended while ed holds " + filepath.Base(f) +
			" unsaved: save it, then close the window"
		t.run("display-message", "-t", win, say)
		return nil
	}
	out, err := t.run("list-panes", "-a", "-f", "#{==:#{@frame},"+win+"}",
		"-F", "#{pane_id} #{window_id} #{pane_dead} #{@slot}"+
			" #{pane_current_command} #{pane_start_command}")
	if err != nil {
		return nil
	}
	sh, _ := t.run("show", "-gv", "default-shell")
	here, _ := t.run("show", "-wqv", "-t", win, "@studio-here")
	panel, _ := t.run("show", "-wqv", "-t", win, "@studio-panel")
	var ends []string
	stays, home := false, true
	for _, l := range strings.Split(out, "\n") {
		f := strings.SplitN(l, " ", 6)
		if len(f) == 5 { // a shell's start command, none, trimmed off
			f = append(f, "")
		}
		if len(f) < 6 {
			continue
		}
		pane, at, dead, slot, cmd, start := f[0], f[1], f[2], f[3],
			f[4], f[5]
		// a slot's program was started in its pane, and a shell was
		// not: one at its prompt runs nothing but itself
		if dead == "1" || pane != here && (cmd == filepath.Base(sh) ||
			start != "" && slot != panel) {
			ends = append(ends, pane)
			continue
		}
		stays = true
		if at != win {
			_, err := t.run("join-pane", "-d", "-s", pane, "-t",
				win)
			home = home && err == nil
		}
	}
	for _, pane := range ends {
		t.run("kill-pane", "-t", pane) // a kill runs no hook
	}
	if home {
		t.run("kill-session", "-t", "="+back(win))
	}
	if stays {
		plain(t, win)
	}
	return nil
}
