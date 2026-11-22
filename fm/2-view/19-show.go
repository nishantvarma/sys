// show draws the screen, with word on the left of the bottom line.
func (v *term) show(word string) {
	w, h := v.t.Size()
	v.t.Write(v.screen(word, w, h))
	v.t.Flush()
}
