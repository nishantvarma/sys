// spawn runs a child on our terminal, suspending the UI for its duration.
// Off the terminal, in a window or a stream, only o and e run, for they
// open what they are given beside fm, through `sh/spawn`, and return; they
// run with nothing to read or write, and what else needs the terminal says
// so.
func (m *fm) spawn(cur, wait bool, args ...string) {
	if cur {
		args = append(args, m.cur())
	}
	var cmd []string
	for _, a := range args {
		if a != "" {
			cmd = append(cmd, a)
		}
	}
	if len(cmd) == 0 {
		return
	}
	if m.t == nil && cmd[0] != cmdOpen && cmd[0] != cmdEdit {
		m.flash(cmd[0] + ": needs the terminal")
		return
	}
	if m.t == nil {
		c := exec.Command(cmd[0], cmd[1:]...)
		c.Dir = m.cwd
		if err := c.Start(); err != nil {
			m.flash(cmd[0] + ": " + err.Error())
			return
		}
		go c.Wait() // reap it, as detach does
		return
	}
	m.t.Suspend()
	defer m.t.Resume()
	tookInterrupt() // stale Ctrl-C; the child owns the next one
	c := exec.Command(cmd[0], cmd[1:]...)
	c.Dir = m.cwd
	c.Stdin, c.Stdout, c.Stderr = os.Stdin, os.Stdout, os.Stderr
	err := c.Run()
	if tookInterrupt() {
		return // Ctrl-C: back to the pane, with nothing to read
	}
	// ran: started but exited non-zero; bad: never started at all
	// (not found, not executable).
	_, ran := err.(*exec.ExitError)
	bad := err != nil && !ran
	if bad {
		fmt.Fprintln(os.Stderr, "fm:", err)
	}
	if wait || ran || bad {
		p := exec.Command("pause")
		p.Stdin, p.Stdout, p.Stderr = os.Stdin, os.Stdout, os.Stderr
		p.Run()
	}
}
