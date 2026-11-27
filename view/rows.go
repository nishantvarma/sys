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

// curRow is the index of the cursor's row among rows: the one a move
// went to, else the one the program drew it on, else the first.
func (v *Term) curRow(rows []*outline.Node) int {
	for i, r := range rows {
		if v.isCur(r) {
			return i
		}
	}
	return 0
}

// isCur is whether the cursor is on row n: the one a move went to, till
// the program draws again, else the one it drew it on.
func (v *Term) isCur(n *outline.Node) bool {
	if v.cur != "" {
		return n.ID == v.cur
	}
	return n.Cur
}
