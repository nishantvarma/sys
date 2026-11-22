// draw says what to show, as an outline: the buffer, a node for each line;
// and the status, ed's word and its state, the mode and + while unsaved.
// How it looks, and which lines show, are the view's.
func (e *editor) draw() {
	a, c := e.rng()
	mode := e.mode.name()
	if e.ext {
		mode = "extend"
	}
	state := []string{mode}
	if e.b.dirty {
		state = append(state, "+")
	}
	n := &outline.Node{Kind: "text", Text: e.b.path}
	for i := range e.b.rows() {
		n.Kids = append(n.Kids, e.lineNode(i, a, c))
	}
	e.ui.Draw(n, e.msg, state)
}
