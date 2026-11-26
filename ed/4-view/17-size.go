// Size is how much of the text the view shows at once: the terminal, but
// the bar's line (tops), so that a page is what a screen holds.
func (v *term) Size() (int, int) {
	w, h := v.t.Size()
	return w, h - tops(h)
}
