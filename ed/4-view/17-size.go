// Size is how much of the text the view shows at once: the terminal.
func (v *term) Size() (int, int) {
	return v.t.Size()
}
