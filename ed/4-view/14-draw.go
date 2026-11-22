func (v *term) Draw(n *outline.Node, status string, state []string) {
	w, h := v.t.Size()
	v.t.Write(v.render(n, status, state, w, h))
	v.t.Flush()
}
