// Draw shows n, the text, as ed's own screen (render).
func (v *term) Draw(n *outline.Node, status string, state []string) {
	w, h := v.t.Size()
	v.last, v.mode = n, state[0]
	v.t.Frame(v.render(n, status, state, w, h))
	if v.ended {
		v.ended = false
		// off X there is none to set
		tty.Put(tty.Primary, outline.Selected(n))
	}
}
