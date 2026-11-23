// open lays the frame out in a new tmux window in dir: its pane split
// down among the slots, theirs across, and so on; then each slot starts
// what it runs. First the window's own hooks are emptied, for a hook
// that lays a window out after a split or a pane's end, as `tmux/conf`'s
// do (`tmux/conf:/select-layout main-vertical/`), would undo the frame.
// Its borders are drawn in the rule's colour, the palette's, as the
// samples drew the frame. The window keeps each slot's pane by name,
// @studio-name, and the frame it was laid out from, @studio-frame, and it
// becomes the studio a program in another window reaches, @studio.
func (f *frame) open(dir string, dry bool) error {
	t := &tmux{dry: dry}
	top, err := t.run("new-window", "-P", "-F", "#{pane_id}", "-c", dir,
		"-n", "studio")
	if err != nil {
		return err
	}
	win, err := t.run("display", "-p", "-t", top, "#{window_id}")
	if err != nil {
		return err
	}
	for _, h := range []string{"after-split-window", "after-kill-pane",
		"pane-exited"} {
		if _, err := t.run("set-hook", "-w", "-t", win, h,
			""); err != nil {
			return err
		}
	}
	panes := map[*slot]string{}
	if err := t.carve(top, f.slots, false, dir, panes); err != nil {
		return err
	}
	var cmds [][]string
	for _, s := range leaves(f.slots) {
		if s.cmd != "" {
			run := []string{"respawn-pane", "-k", "-t", panes[s],
				"-c", dir, s.cmd}
			cmds = append(cmds, run)
		}
		cmds = append(cmds, []string{"set", "-w", "-t", win,
			"@studio-" + s.name, panes[s]})
	}
	cmds = append(cmds,
		[]string{"set", "-w", "-t", win, "pane-border-style",
			"fg=" + tty.Rule},
		[]string{"set", "-w", "-t", win, "pane-active-border-style",
			"fg=" + tty.Rule},
		[]string{"set", "-w", "-t", win, "@studio-frame", f.note},
		[]string{"set", "-g", "@studio", win},
		[]string{"select-pane", "-t", panes[leaves(f.slots)[0]]})
	for _, c := range cmds {
		if _, err := t.run(c...); err != nil {
			return err
		}
	}
	return nil
}
