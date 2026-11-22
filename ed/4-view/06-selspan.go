// selSpan returns the inclusive rune range of line ln under the selection,
// read from its runs, and whether its line break is selected. lo == -1
// means no rune is.
func selSpan(ln *outline.Node) (lo, hi int, tail bool) {
	lo, hi = -1, -1
	i := 0
	for _, r := range ln.Kids {
		if r.Text == "\n" {
			tail = r.Sel
			continue
		}
		k := len([]rune(r.Text))
		if r.Sel && k > 0 {
			if lo == -1 {
				lo = i
			}
			hi = i + k - 1
		}
		i += k
	}
	return
}
