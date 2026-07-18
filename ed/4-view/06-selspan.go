// selSpan returns the inclusive rune range of line idx under the
// selection, and whether the trailing line break is selected. lo == -1
// means no selection.
func selSpan(f frame, idx int) (lo, hi int, tail bool) {
	if !f.sel || idx < f.a.line || idx > f.c.line {
		return -1, -1, false
	}
	n := len(f.lines[idx])
	lo, hi, tail = 0, n-1, idx < f.c.line
	if idx == f.a.line {
		lo = f.a.col
	}
	if idx == f.c.line {
		hi = f.c.col
		tail = f.c.col >= n
	}
	if hi >= n {
		hi = n - 1
	}
	return
}
