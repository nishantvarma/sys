// entry is a directory child prepared for display: for a file, its size
// and when it last changed too, from the stat that finds a program.
type entry struct {
	name string
	path string
	dir  bool
	link bool
	bad  bool // symlink with a missing target
	exec bool
	size int64
	mod  time.Time
}
