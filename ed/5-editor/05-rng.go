func (e *editor) rng() (pos, pos) {
	return order(e.caret.anchor, e.caret.head)
}
