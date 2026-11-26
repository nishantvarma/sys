// Draw shows n through the lens shown: source as ed's own screen
// (render), another as sys/view draws it, under the bar (lensed).
func (v *term) Draw(n *outline.Node, status string, state []string) {
	w, h := v.t.Size()
	v.last, v.mode = n, state[0]
	if v.lens > 0 {
		v.t.Frame(v.lensed(n, status, state, w, h), "")
	} else {
		v.t.Frame(v.render(n, status, state, w, h))
	}
	if v.ended {
		v.ended = false
		tty.Put(tty.Primary, selected(n)) // off X there is none to set
	}
}
