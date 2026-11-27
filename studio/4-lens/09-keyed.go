// keyed is a key read in a goroutine of its own, and whether the view
// was still there to read it from.
type keyed struct {
	ev outline.Key
	ok bool
}
