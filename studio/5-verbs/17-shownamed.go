// showNamed shows the tab named name, in whichever slot of the frame's
// it waits.
func (t *tmux) showNamed(win, name string) error {
	out, err := t.run("list-panes", "-a", "-f",
		"#{&&:#{==:#{@frame},"+win+"},#{==:#{@tab},"+name+"}}",
		"-F", "#{pane_id} #{window_id} #{@slot}")
	if err != nil {
		return err
	}
	f := strings.Fields(out)
	if len(f) < 3 {
		return fmt.Errorf("no tab %s in window %s", name, win)
	}
	return t.show(win, f[2], tabPane{pane: f[0], window: f[1],
		name: name})
}
