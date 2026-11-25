package view

import "sys/outline"

// rows are what the view lays out: the kids of the tree drawn last, a row
// or a card each; none before the first draw.
func (v *Term) rows() []*outline.Node {
	if v.last == nil {
		return nil
	}
	return v.last.Kids
}

// curRow is the cursor's row, else the first.
func curRow(rows []*outline.Node) int {
	for i, r := range rows {
		if r.Cur {
			return i
		}
	}
	return 0
}
