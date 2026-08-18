// indent prefixes a tab to every non-empty line the selection touches.
func (e *editor) indent() {
	a, c := e.rng()
	e.b.snapshot(e.caret)
	for l := a.line; l <= c.line; l++ {
		if e.b.lineLen(l) > 0 {
			e.b.mapLine(l, func(ln []rune) []rune {
				return append([]rune{'\t'}, ln...)
			})
		}
	}
	e.place(a)
}
