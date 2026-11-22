// row is a node as a line on screen, and how deep in the tree it sits.
type row struct {
	n     *outline.Node
	depth int
}
