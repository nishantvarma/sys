package view

import (
	"sys/outline"
	"sys/tty"
)

// shut is the one hint a tree draws, dim, as scaffolding is: something is
// inside, not shown.
var shut = tty.Dim("▸") + " "

// row is a node as a line on screen: how deep in the tree it sits, and,
// in the tree, the hint drawn before it (see flat).
type row struct {
	n     *outline.Node
	depth int
	lead  string
}

// flat is the rows a tree shows: n's kids, and under each open one its own,
// a level deeper. Each row's lead is its hint, and only what hides
// something has one: shut, on a folder or a node with kids that is not
// open. An open node's kids, under it, say it is open, so it has spaces,
// as a leaf has, and the names line up; depth is the indent alone. A text
// file's items stand in a bracket, the left side of a box (see bracket).
// The hints are the view's; the program sends the same tree.
func flat(n *outline.Node, depth int, indent string) []row {
	var rows []row
	for _, k := range n.Kids {
		mark := "  "
		if !k.Open && (k.Kind == "dir" || len(k.Kids) > 0) {
			mark = shut
		}
		rows = append(rows, row{k, depth, indent + mark})
		if k.Open {
			sub := flat(k, depth+1, indent+"  ")
			items := len(k.Kids) > 0 && k.Kids[0].Kind == "item"
			if items && k.Kind != "item" { // a file's items
				bracket(sub, len(indent))
			}
			rows = append(rows, sub...)
		}
	}
	return rows
}

// bracket draws the left side of a box down the rows of a file's items, in
// the indent that starts at byte at of each lead: ╭ on the first row, │ on
// those between and ╰ on the last, so a file's insides read apart from the
// folders around them.
func bracket(rows []row, at int) {
	for i := range rows {
		g := "│ "
		switch {
		case i == len(rows)-1:
			g = "╰ "
		case i == 0:
			g = "╭ "
		}
		l := rows[i].lead
		rows[i].lead = l[:at] + g + l[at+len("  "):]
	}
}

// zoom is the node the cursor is among the kids of: the folder a zoomed
// view shows. nil when no node says so.
func zoom(n *outline.Node) *outline.Node {
	if n.In {
		return n
	}
	for _, k := range n.Kids {
		if z := zoom(k); z != nil {
			return z
		}
	}
	return nil
}

// shown is what the view shows of the tree drawn last: as the folder or as
// tiles, the cursor's folder and its rows; as the tree, every open row from
// the top, under the top's path.
func (v *Term) shown() (*outline.Node, []row) {
	if v.lay == "tree" {
		return v.last, flat(v.last, 0, "")
	}
	z := zoom(v.last)
	if z == nil {
		z = v.last
	}
	rows := make([]row, len(z.Kids))
	for i, k := range z.Kids {
		rows[i] = row{n: k}
	}
	return z, rows
}

// curRow is the cursor's row: the one it is on, else the folder it is in,
// as an empty folder holds it, else the first.
func curRow(rows []row) int {
	for i, r := range rows {
		if r.n.Cur {
			return i
		}
	}
	for i, r := range rows {
		if r.n.In {
			return i
		}
	}
	return 0
}
