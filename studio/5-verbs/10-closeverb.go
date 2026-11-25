// closeVerb is what a frame's window runs as a pane of it ends (open). When
// the pane was the frame's face, fm's slot (face), the frame ends with
// it, at once: each slot's program is killed, and each shell waiting at
// its prompt. A shell running a command stays, and so does the shell the
// frame was laid out around, @studio-here; the window is then a plain one
// again (plain). But while ed names a file it holds unsaved, which fm's
// second q takes away as it lets the work go (`fm/4-main/40-quit.go`),
// the face ended some other way, as a crash ends it, and the frame stays
// for ed to save, saying so. Any other pane's end is its own: a shell
// that exits ends alone.
func closeVerb(args []string) error {
	if len(args) != 2 {
		return errors.New("usage: studio close window pane")
	}
	win, ended := args[0], args[1]
	t := &tmux{}
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
	out, err := t.run("list-panes", "-t", win, "-F",
		"#{pane_id} #{pane_current_command} #{pane_start_command}")
	if err != nil {
		return nil
	}
	sh, _ := t.run("show", "-gv", "default-shell")
	here, _ := t.run("show", "-wqv", "-t", win, "@studio-here")
	stays := false
	for _, l := range strings.Split(out, "\n") {
		pane, rest, _ := strings.Cut(l, " ")
		cmd, start, _ := strings.Cut(rest, " ")
		// a slot's program was started in its pane, and a shell was
		// not: one at its prompt runs nothing but itself
		if pane != here && (start != "" || cmd == filepath.Base(sh)) {
			t.run("kill-pane", "-t", pane) // a kill runs no hook
		} else {
			stays = true
		}
	}
	if stays {
		plain(t, win)
	}
	return nil
}
