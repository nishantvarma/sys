// stream is fm with its view at the other end of its own input and
// output, a JSON line each way (outline.Stream): a host starts fm with
// pipes and draws it where it likes, as studio will once it hosts the
// slots (`todo/03-studio/01-slots`). Its output is the
// view's, so nothing else of fm's is written there; what needs a terminal
// says so, as in a window. It starts without the eye, saying nothing to
// cur, until the host's focus key gives it.
func (m *fm) stream() error {
	if err := os.Chdir(m.cwd); err != nil {
		return err
	}
	m.ui = outline.Stream(os.Stdin, os.Stdout)
	m.on = false
	m.ls()
	m.land(m.pos[m.cwd])
	m.loop()
	return nil
}
