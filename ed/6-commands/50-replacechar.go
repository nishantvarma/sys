// replaceChar asks for one key and overwrites every rune the selection
// covers with it, line breaks untouched. The selection stays: its text
// is still there.
func (e *editor) replaceChar() {
	k, ok := e.ask("replace:")
	if !ok || k.Name != "" || !unicode.IsPrint(k.Rune) {
		return
	}
	a, c := e.rng()
	e.b.snapshot(e.caret)
	for l := a.line; l <= c.line; l++ {
		lo, hi := 0, e.b.lineLen(l)
		if l == a.line {
			lo = a.col
		}
		if l == c.line {
			hi = min(c.col+1, hi)
		}
		e.b.mapLine(l, func(ln []rune) []rune {
			for i := lo; i < hi; i++ {
				ln[i] = k.Rune
			}
			return ln
		})
	}
}
