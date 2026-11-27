// open lays the frame out in dir: in a new tmux window, or, given here, in
// that pane's window, where it is a shell alone and stays, the frame's
// shell (shell). The window's pane is split down among the slots, theirs
// across, and so on; then each slot starts what it runs, told its slot,
// $STUDIO (state.Slot). First the window's own hooks are set: its layout
// changing does nothing, for `tmux/conf:/window-layout-changed/` lays a
// window out again when its panes change in number, which would undo the
// frame; and a pane's end runs closeVerb, which ends the frame when the
// pane was its face, fm's slot (face).
// Its borders are drawn in the rule's colour, the palette's, as the
// samples drew the frame. Each slot's pane carries the slot's name,
// @slot, which is how e, o and p -a find it, asking tmux and not studio
// (`design/mouse`), the name of the tab shown in it, @tab, and its
// frame's window, @frame. A slot of tabs shows its first, and the rest
// wait out of sight (park), the slot's bar naming them (bars); a tab of
// one stays, dead, as what runs in it ends, for the next to take its
// place (closeVerb). The window keeps the frame it was laid out from,
// @studio-frame, its face, @studio-face, and the shell it was laid out
// around, @studio-here; and it becomes the studio a program in another
// window reaches, @studio.
func (f *frame) open(dir, here string, dry bool) error {
	t := &tmux{dry: dry}
	top, err := here, error(nil)
	if top == "" {
		top, err = t.run("new-window", "-P", "-F", "#{pane_id}", "-c",
			dir, "-n", "studio")
		if err != nil {
			return err
		}
	}
	win, err := t.run("display", "-p", "-t", top, "#{window_id}")
	if err != nil {
		return err
	}
	self, err := os.Executable()
	if err != nil {
		return err
	}
	end := fmt.Sprintf("run-shell -b \"'%s' close %s #{hook_pane}\"",
		self, win)
	for _, h := range hooks(end) {
		if _, err := t.run("set-hook", "-w", "-t", win, h[0],
			h[1]); err != nil {
			return err
		}
	}
	panes := map[*slot]string{}
	if err := t.carve(top, f.slots, false, dir, panes); err != nil {
		return err
	}
	first := leaves(f.slots)[0]
	var cmds [][]string
	if here != "" {
		// carve leaves the shell that was here in the first slot: it
		// trades places with the pane made for the frame's shell
		if sh := f.shell(); sh != first {
			cmds = append(cmds, []string{"swap-pane", "-d", "-s",
				here, "-t", panes[sh]})
			panes[first], panes[sh] = panes[sh], here
		}
		cmds = append(cmds, []string{"set", "-w", "-t", win,
			"@studio-here", here})
	}
	var tabbed []*slot
	for _, s := range leaves(f.slots) {
		if s.cmd != "" {
			run := []string{"respawn-pane", "-k", "-t", panes[s],
				"-e", "STUDIO=" + s.name, "-c", dir, s.cmd}
			cmds = append(cmds, run)
		}
		for _, o := range [][2]string{{"@slot", s.name},
			{"@tab", s.tabs[0].name}, {"@frame", win}} {
			cmds = append(cmds, []string{"set", "-p", "-t",
				panes[s], o[0], o[1]})
		}
		if len(s.tabs) > 1 {
			tabbed = append(tabbed, s)
			cmds = append(cmds, []string{"set", "-p", "-t",
				panes[s], "remain-on-exit", "on"})
		}
	}
	for _, o := range borders {
		cmds = append(cmds, []string{"set", "-w", "-t", win, o,
			"fg=" + tty.Rule})
	}
	if sh := f.shell(); sh != nil {
		cmds = append(cmds, []string{"set", "-w", "-t", win,
			"@studio-panel", sh.name})
	}
	cmds = append(cmds,
		[]string{"set", "-w", "-t", win, "@studio-frame", f.note},
		[]string{"set", "-g", "@studio", win},
		[]string{"select-pane", "-t", panes[first]})
	if fc := f.face(); fc != nil {
		cmds = append(cmds, []string{"set", "-w", "-t", win,
			"@studio-face", panes[fc]})
	}
	for _, c := range cmds {
		if _, err := t.run(c...); err != nil {
			return err
		}
	}
	for _, s := range tabbed {
		for _, tb := range s.tabs[1:] {
			if _, err := t.park(win, dir, s.name, tb.name, tb.cmd,
				told(s.name, win, tb.cmd)); err != nil {
				return err
			}
		}
		if err := t.bars(win, s.name); err != nil {
			return err
		}
	}
	return nil
}
