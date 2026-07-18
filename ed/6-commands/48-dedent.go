// dedent strips one leading tab or space from every line the
// selection touches.
func (e *editor) dedent() {
	a, c := e.rng()
	e.b.snapshot()
	for l := a.line; l <= c.line; l++ {
		ln := e.b.lines[l]
		if len(ln) > 0 && (ln[0] == '\t' || ln[0] == ' ') {
			e.b.lines[l] = ln[1:]
		}
	}
	e.b.dirty = true
	e.place(a)
}
