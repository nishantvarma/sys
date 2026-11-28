// lineNode is line i as a node: its text, and when the selection or the
// cursor falls in it, runs that mark them. a and c are the selection's
// ordered bounds, inclusive; a selected line break is a run of its own,
// "\n".
func (e *editor) lineNode(i int, a, c pos) *outline.Node {
	line := e.b.line(i)
	n := &outline.Node{Kind: "line", Text: string(line)}
	head := e.head()
	sel := a != c && i >= a.line && i <= c.line
	if !sel && i != head.line {
		return n
	}
	// the runes under the selection; lo > hi when only the break is
	lo, hi, tail := -1, -1, false
	if sel {
		lo, hi, tail = 0, len(line)-1, i < c.line
		if i == a.line {
			lo = a.col
		}
		if i == c.line {
			hi = c.col
			tail = c.col >= len(line)
		}
		hi = min(hi, len(line)-1)
	}
	at := -1 // the cursor's column, on its own line
	if i == head.line {
		at = min(head.col, len(line))
	}
	var cuts []int
	for _, x := range []int{lo, hi + 1, at} {
		if x > 0 && x < len(line) && (x == at || lo <= hi) {
			cuts = append(cuts, x)
		}
	}
	slices.Sort(cuts)
	start := 0
	for _, end := range append(slices.Compact(cuts), len(line)) {
		if end > start {
			in := lo >= 0 && lo <= hi && start >= lo && end-1 <= hi
			n.Runs = append(n.Runs, &outline.Node{
				Text: string(line[start:end]),
				Sel:  in,
				Cur:  start == at,
			})
		}
		start = end
	}
	if at == len(line) {
		n.Runs = append(n.Runs, &outline.Node{Cur: true})
	}
	if tail {
		n.Runs = append(n.Runs, &outline.Node{Text: "\n", Sel: true})
	}
	return n
}
