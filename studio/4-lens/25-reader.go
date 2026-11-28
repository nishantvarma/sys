// reader is the right tool for a kind of thing
// (`design/lens:/The reader, by kind/`): the suffix of the names it reads,
// "" for a note, a name with no dot; what it calls the thing and what it
// reads it by, said as the lens's word; and the reading, a thing's text
// into its parts, each named by the id given and its line; as far as it
// can read, when it says why it stopped.
type reader struct {
	suffix, thing, by string
	read              func(r io.Reader, id string) ([]*outline.Node, error)
}
