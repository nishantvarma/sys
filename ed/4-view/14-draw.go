// Draw shows n as ed says it: the text, source, as ed's own screen
// (render), and any other lens, a tree of items, as sys/view draws it,
// under the bar (lensed).
func (v *term) Draw(n *outline.Node, status string, state []string) {
	w, h := v.t.Size()
	v.last, v.state, v.mode = n, state, state[0]
	if n.Kind != "text" {
		v.t.Frame(v.lensed(n, status, state, w, h), "")
	} else {
		v.t.Frame(v.render(n, status, state, w, h))
	}
	if v.ended {
		v.ended = false
		tty.Put(tty.Primary, selected(n)) // off X there is none to set
	}
}
