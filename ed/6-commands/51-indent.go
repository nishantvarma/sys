// indent prefixes a tab to every non-empty line the selection touches.
// The selection follows its text, so > > indents twice.
func (e *editor) indent() {
	a, c := e.rng()
	e.b.snapshot(e.caret)
	for l := a.line; l <= c.line; l++ {
		if e.b.lineLen(l) == 0 {
			continue
		}
		e.b.mapLine(l, func(ln []rune) []rune {
			return append([]rune{'\t'}, ln...)
		})
		e.follow(l, 1)
	}
}
