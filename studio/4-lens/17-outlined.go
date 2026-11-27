// outlined is the thing's outline: its note's items by their indent
// (outline.Text), a row an item, two spaces in for each item it is under,
// each named by its line, so that a key on a row is a key on that line.
func outlined(th thing) ([]*outline.Node, error) {
	if th.note == "" {
		return nil, th.why
	}
	var rows []*outline.Node
	var walk func(ns []*outline.Node, depth int)
	walk = func(ns []*outline.Node, depth int) {
		for _, it := range ns {
			rows = append(rows, &outline.Node{ID: it.ID,
				Kind: it.Kind,
				Text: strings.Repeat("  ", depth) + it.Text})
			walk(it.Kids, depth+1)
		}
	}
	walk(outline.Text(bytes.NewReader(th.text), th.note), 0)
	return rows, nil
}
