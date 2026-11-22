// place is where an item is: the text file it is in, the node above it,
// an item or the file's row, and its first kid, when it has any.
type place struct {
	file, up, first string
}
