// frame is an immutable snapshot of what to paint.
type frame struct {
	lines [][]rune
	a, c  pos // ordered selection bounds (inclusive)
	head  pos // cursor
	sel   bool
	top   int
	off   int
	w, h  int
	mode  string
	path  string
	dirty bool
	msg   string
}
