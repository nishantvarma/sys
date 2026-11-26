// lensKey is what k does while a lens other than source shows, where
// nothing is typed: Tab turns to the next lens, esc and enter go back to
// source, on the line ed's cursor is on, where a move among the items put
// it, and ctrl-s saves, as in every mode. Any other key is nothing here:
// a view's own, a move or t, it keeps and does, and so none reaches the
// text out of sight.
func (e *editor) lensKey(k tty.Key) {
	switch k.Token() {
	case tty.KTab:
		e.turnLens()
	case tty.KEsc, tty.KEnter:
		e.lens = 0
	case "C-s":
		e.doSave()
	}
}
