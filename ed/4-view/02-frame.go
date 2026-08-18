// frame is an immutable snapshot of what to paint. It asks for lines
// rather than holding them: a count, and a way to fetch one. Whatever the
// buffer keeps its text in, the view only ever needs row idx as runes, so
// this is where the two stop having to agree.
type frame struct {
	rows  int              // lines in the buffer, not rows on screen
	line  func(int) []rune // row idx, read-only
	a, c  pos              // ordered selection bounds (inclusive)
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
