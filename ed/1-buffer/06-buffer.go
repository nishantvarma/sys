type buffer struct {
	lines [][]rune
	path  string
	disk  [32]byte // sum of the file as ed last read or wrote it
	dirty bool
	undo  []snap
	redo  []snap
	edits int // snapshots taken: one per edit begun
}
