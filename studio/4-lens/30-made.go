// made is a tree a typed lens made of a thing, how it saw it, and what it
// made it from: the lens, the thing's path, its note and that note's
// text. While those stay as they were, the tree is the same, and a draw
// reads nothing again.
type made struct {
	lens       int
	path, note string
	text       []byte
	tree       *outline.Node
	word       string
}
