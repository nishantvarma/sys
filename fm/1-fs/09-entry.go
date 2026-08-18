// entry is a directory child prepared for display.
type entry struct {
	name string
	path string
	dir  bool
	link bool
	bad  bool // symlink with a missing target
	exec bool
}
