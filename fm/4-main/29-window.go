// window is fm in a window, its view the program window at the other end
// of a pipe (outline.Window). What needs a terminal says so instead.
func (m *fm) window() error {
	if err := os.Chdir(m.cwd); err != nil {
		return err
	}
	ui, done, err := outline.Window()
	if err != nil {
		return err
	}
	defer done()
	m.ui = ui
	m.ls()
	m.land(m.pos[m.cwd])
	m.loop()
	return nil
}
