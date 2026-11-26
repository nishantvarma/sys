// draw says what to show, as an outline: the buffer, a node for each line,
// named by its address, path:line, so that a key or a click names one, as
// a lens names it; and the status, ed's word and its state: the mode, +
// while unsaved, and held while that keeps ed from cur's file (chase).
// The lenses ride on the tree's tags, for a view's bar, and a lens other
// than source is drawn as that lens takes the text (lensed), its name in
// the state. How it looks, and which lines show, are the view's.
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
	if e.held {
		state = append(state, "held")
	}
	at := e.where()
	n := &outline.Node{ID: at, Kind: "text", Text: e.b.path, Tags: lenses}
	for i := range e.b.rows() {
		line := e.lineNode(i, a, c)
		line.ID = outline.At(at, i+1)
		n.Kids = append(n.Kids, line)
	}
	if e.lens > 0 {
		n = e.lensed(n)
		state = append(state, lenses[e.lens])
	}
	e.ui.Draw(n, e.msg, state)
}
