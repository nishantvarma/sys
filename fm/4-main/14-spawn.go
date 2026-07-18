// spawn runs a child on our terminal, suspending the UI for its duration.
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
	m.t.Suspend()
	defer m.t.Resume()
	c := exec.Command(cmd[0], cmd[1:]...)
	c.Dir = m.ctx()
	c.Stdin, c.Stdout, c.Stderr = os.Stdin, os.Stdout, os.Stderr
	err := c.Run()
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
