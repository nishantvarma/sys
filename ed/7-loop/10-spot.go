// spot is where a key on a line names: its node's id, path:line as draw
// names it, at the rune the key says, else the line's start. false for a
// line of another file, or no line.
func (e *editor) spot(id string, at int) (pos, bool) {
	path, n := outline.Address(id)
	if n < 1 || path != e.where() {
		return pos{}, false
	}
	return pos{n - 1, max(0, at-1)}, true
}
