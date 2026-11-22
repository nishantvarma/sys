// header is the top line: where the view is, a path or an item's address,
// cut from the left when long, for its end names the place.
func header(n *outline.Node, w int) string {
	return tty.Bold(tty.ClipLeft(tilde(n.ID), w))
}
