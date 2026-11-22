// window is fm in a window: its view is the program window, at the other
// end of a pipe, sent the outline and sending back keys, a JSON line each.
// What needs a terminal says so instead.
func (m *fm) window() error {
	if err := os.Chdir(m.cwd); err != nil {
		return err
	}
	c := exec.Command("window")
	in, err := c.StdinPipe()
	if err != nil {
		return err
	}
	out, err := c.StdoutPipe()
	if err != nil {
		return err
	}
	c.Stderr = os.Stderr
	if err := c.Start(); err != nil {
		return err
	}
	defer c.Wait()
	defer in.Close() // its input ends, and so does the window
	m.ui = outline.Stream(out, in)
	m.ls()
	m.land(m.pos[m.cwd])
	m.loop()
	return nil
}
