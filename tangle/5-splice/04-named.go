// named is what a chunk's name can be, a path's parts without their
// numbers: a ref to anything else is text.
var named = regexp.MustCompile(`^[A-Za-z0-9_./-]+$`)
