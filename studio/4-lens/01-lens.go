// lens is one way to see the thing at cur (`design/lens`): its name, which
// the bar shows and studio show takes; what it makes of the thing it is on
// (thing), a tree whose kind a view draws it by, and how it saw it, said
// as the program's word; and whether it reads the note's text, ed's while
// ed holds it unsaved, or the file as saved, as git does.
type lens struct {
	name  string
	nodes func(t thing) (*outline.Node, string, error)
	typed bool
}
