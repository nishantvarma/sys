// row is a node as a line on screen: how deep in the tree it sits, and,
// in the tree, the hint drawn before it (see flat).
type row struct {
	n     *outline.Node
	depth int
	lead  string
}
