// replaceChar reads one key and overwrites every rune the selection covers
// with it, line breaks untouched — the read-one-key pattern findChar uses.
func (e *editor) replaceChar() {
	e.msg = "replace:"
	e.draw()
	k, ok := e.t.ReadKey()
	e.msg = ""
	if !ok || k.Name != "" || !unicode.IsPrint(k.Rune) {
		return
	}
	a, c := e.rng()
	e.b.snapshot()
	for l := a.line; l <= c.line; l++ {
		ln := e.b.lines[l]
		lo, hi := 0, len(ln)
		if l == a.line {
			lo = a.col
		}
		if l == c.line {
			hi = min(c.col+1, len(ln))
		}
		for i := lo; i < hi; i++ {
			ln[i] = k.Rune
		}
	}
	e.b.dirty = true
	e.place(a)
}
