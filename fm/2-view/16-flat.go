// flat is the rows a tree shows: n's kids, and under each open one its own,
// a level deeper. Each row's lead is its hint: ⌄ on an open node, › on one
// that can open, a folder or a node with more inside, and spaces on the
// rest, so that the names line up; depth is the indent alone. A text
// file's items stand in a bracket, the left side of a box (see bracket).
// The hints are the view's; fm sends the same tree.
func flat(n *outline.Node, depth int, indent string) []row {
	var rows []row
	for _, k := range n.Kids {
		mark := "  "
		switch {
		case k.Open:
			mark = "⌄ "
		case k.Kind == "dir" || len(k.Kids) > 0:
			mark = "› "
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
