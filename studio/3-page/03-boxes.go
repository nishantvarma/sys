// boxes writes a split as the page lays it: a column down, a row across,
// each slot as tall or as wide as its share, and a slot of slots split
// the other way inside it, as carve does in the terminal.
func boxes(b *strings.Builder, kids []*slot, across bool) {
	way := "down"
	if across {
		way = "across"
	}
	fmt.Fprintf(b, "<div class=\"%s\">\n", way)
	for i, w := range weights(kids) {
		k := kids[i]
		fmt.Fprintf(b, "<div class=\"slot\" style=\"flex: %d 1 0\">\n",
			w)
		if len(k.kids) > 0 {
			boxes(b, k.kids, !across)
		} else {
			leaf(b, k)
		}
		b.WriteString("</div>\n")
	}
	b.WriteString("</div>\n")
}
