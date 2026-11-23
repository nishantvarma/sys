// reach puts the cursor where a key named: its node's id, path:line, as
// draw names it, at the rune the key says, else at the line's start; one
// past the end is the last. false for a line of another file, or no line.
func (e *editor) reach(id string, at int) bool {
	path, n := outline.Address(id)
	if n < 1 || path != e.where() {
		return false
	}
	e.place(pos{n - 1, max(0, at-1)})
	return true
}
