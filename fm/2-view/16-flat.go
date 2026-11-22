// flat is the rows a tree shows: n's kids, and under each open one its own,
// a level deeper.
func flat(n *outline.Node, depth int) []row {
	var rows []row
	for _, k := range n.Kids {
		rows = append(rows, row{k, depth})
		if k.Open {
			rows = append(rows, flat(k, depth+1)...)
		}
	}
	return rows
}
