package view

import "sys/outline"

// row is a node as a line on screen: how deep in the tree it sits, the
// lead before it, its indent with any bracket in it (see flat), its hint
// (see hint), and whether it is one of a note's own top items, as a
// note's title is.
type row struct {
	n     *outline.Node
	depth int
	lead  string
	hint  string
	first bool
}

// hint is the view's mark on a node: + when something is inside and not
// shown, - when it is open and its insides show under it, and a blank on
// a leaf. Plain signs, so that every font draws them alike, and dim, as
// scaffolding is. The hints are the view's; the program sends the tree.
func hint(n *outline.Node) string {
	inside := n.Kind == "dir" || len(n.Kids) > 0
	switch {
	case inside && n.Open:
		return "-"
	case inside:
		return "+"
	}
	return " "
}

// flat is the rows a tree shows: n's kids, and under each open one its own,
// a level deeper, each row's lead its indent. A text file's items stand in
// a bracket, the left side of a box (see bracket).
func flat(n *outline.Node, depth int, indent string) []row {
	var rows []row
	for _, k := range n.Kids {
		first := k.Kind == "item" && n.Kind != "item"
		rows = append(rows, row{k, depth, indent, hint(k), first})
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
// thumbs, the cursor's folder and its rows; as the tree, every open row
// from the top, under the top's path.
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
		rows[i] = row{n: k, hint: hint(k),
			first: k.Kind == "item" && z.Kind != "item"}
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
