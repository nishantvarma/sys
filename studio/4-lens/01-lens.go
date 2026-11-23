// lens is one way to see the thing at cur. A lens with no nodes yet is a
// placeholder: it says what it will show, and what it waits on.
type lens struct {
	name string // what it is called, and set by: studio lens figure
	by   string // how it sees the thing, said as the program's word
	// nodes is the thing's kids as the lens sees them, given its path;
	// an error says why it has none to show
	nodes func(path string) ([]*outline.Node, error)
	kind  string   // what the thing is, seen so, which a view draws by
	waits []string // a placeholder's words, in place of nodes
}
