// lens is one way to see the thing at cur (`design/lens`): its name, which
// the bar shows and studio lens sets; how it sees the thing, said as the
// program's word; the thing's kids as it sees them, from what the lens is
// on (thing); the kind of the tree it makes, which a view draws by; and
// whether it reads the note's text, ed's while ed holds it unsaved, or
// the file as saved, as git does.
type lens struct {
	name  string
	by    string
	nodes func(t thing) ([]*outline.Node, error)
	kind  string
	typed bool
}
