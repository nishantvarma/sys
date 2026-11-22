// screen is the whole screen, w by h: the tree drawn last, laid out as the
// view lays it out, over the bottom line with word on its left.
func (v *term) screen(word string, w, h int) string {
	head, rows := v.shown()
	line := foot(word, v.state, v.words(), w)
	if v.lay == "tiles" {
		return tiled(head, rows, line, w, h)
	}
	return render(head, rows, line, w, h)
}
