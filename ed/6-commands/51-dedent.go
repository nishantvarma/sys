// dedent strips one leading tab or space from every line the
// selection touches.
func (e *editor) dedent() {
	a, c := e.rng()
	e.b.snapshot(e.caret)
	for l := a.line; l <= c.line; l++ {
		ln := e.b.line(l)
		if len(ln) == 0 || (ln[0] != '\t' && ln[0] != ' ') {
			continue
		}
		e.b.mapLine(l, func(ln []rune) []rune { return ln[1:] })
	}
	e.place(a)
}
