// tallyOf is what a lens's tree says of itself in a word: a diff's lines
// added and removed, +2 −1, else how many nodes it holds; nothing when
// the lens only says why it has none.
func tallyOf(t *outline.Node) string {
	add, del, n := 0, 0, 0
	var walk func(ns []*outline.Node)
	walk = func(ns []*outline.Node) {
		for _, k := range ns {
			switch k.Kind {
			case "add":
				add++
			case "del":
				del++
			}
			n++
			walk(k.Kids)
		}
	}
	walk(t.Kids)
	switch {
	case len(t.Kids) > 0 && t.Kids[0].Kind == "waits":
		return ""
	case add+del > 0:
		return fmt.Sprintf("+%d −%d", add, del)
	}
	return strconv.Itoa(n)
}
