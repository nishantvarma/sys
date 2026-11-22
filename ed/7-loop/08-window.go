// window is ed in a window, its view the program window at the other end
// of a pipe (outline.Window); the screen and the keys are the window's.
func (e *editor) window() error {
	ui, done, err := outline.Window()
	if err != nil {
		return err
	}
	defer done()
	e.ui = ui
	e.loop()
	return nil
}
