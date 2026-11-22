// draw says what to show, as an outline: the buffer, a node for each line,
// the mode, whether it is saved, and the status. How it looks, and which
// lines show, are the view's.
func (e *editor) draw() {
	a, c := e.rng()
	mode := e.mode.name()
	if e.ext {
		mode = "extend"
	}
	n := &outline.Node{Kind: "text", Text: e.b.path, Tags: []string{mode}}
	if e.b.dirty {
		n.Tags = append(n.Tags, "+")
	}
	for i := range e.b.rows() {
		n.Kids = append(n.Kids, e.lineNode(i, a, c))
	}
	e.ui.Draw(n, e.msg, nil) // ed says no state yet
}
