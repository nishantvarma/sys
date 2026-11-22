func (v *term) Draw(n *outline.Node, status string, state []string) {
	v.last, v.status, v.state = n, status, state
	v.show(status)
}
