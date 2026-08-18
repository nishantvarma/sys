// snap is one point in history: the text, and the caret that was in it.
// The caret has to ride along — text alone restores to coordinates the
// cursor no longer means, so undo lands a line off wherever the edit moved
// it. lines is this snap's own spine, but the rows it points at belong to
// every other snap too — which is what makes them read-only.
type snap struct {
	lines [][]rune
	caret sel
}
