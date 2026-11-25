// stream is ed with its view at the other end of its own input and
// output, a JSON line each way (outline.Stream): a host starts ed with
// pipes and draws it where it likes, as studio will once it hosts the
// slots (`todo/03-studio/01-slots`). Its output is the
// view's, so nothing else of ed's is written there.
func (e *editor) stream() error {
	e.ui = outline.Stream(os.Stdin, os.Stdout)
	e.loop()
	return nil
}
