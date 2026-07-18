// indent prefixes a tab to every non-empty line the selection touches.
func (e *editor) indent() {
	a, c := e.rng()
	e.b.snapshot()
	for l := a.line; l <= c.line; l++ {
		if len(e.b.lines[l]) > 0 {
			e.b.lines[l] = append([]rune{'\t'}, e.b.lines[l]...)
		}
	}
	e.b.dirty = true
	e.place(a)
}
