// label is how a node reads in a tile: its text and its kind's mark, and …
// when it is shut with more inside.
func label(n *outline.Node) string {
	_, suf := look(n.Kind)
	s := n.Text + suf
	if !n.Open && len(n.Kids) > 0 {
		s += " …"
	}
	return s
}
