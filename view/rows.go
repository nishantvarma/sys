package view

import "sys/outline"

// rows are what the view lays out: the kids of the tree drawn last, a row
// or a card each, and of an outline the rows each part holds, under it,
// in turn, but a folded part's; none before the first draw.
func (v *Term) rows() []*outline.Node {
	if v.last == nil {
		return nil
	}
	if !v.isOutline() {
		return v.last.Kids
	}
	var out []*outline.Node
	var walk func(ns []*outline.Node)
	walk = func(ns []*outline.Node) {
		for _, n := range ns {
			out = append(out, n)
			if !v.folded[n.ID] {
				walk(held(n))
			}
		}
	}
	walk(v.last.Kids)
	return out
}

// curRow is the index of the cursor's row among rows (cursor), else the
// first's.
func (v *Term) curRow(rows []*outline.Node) int {
	c := v.cursor()
	for i, r := range rows {
		if r == c {
			return i
		}
	}
	return 0
}
