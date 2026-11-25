// selected is the text under the selection in the lines n holds, as draw
// marks it: the runs selected, a selected line break a newline.
func selected(n *outline.Node) string {
	var b strings.Builder
	for _, ln := range n.Kids {
		for _, r := range ln.Kids {
			if r.Sel {
				b.WriteString(r.Text)
			}
		}
	}
	return b.String()
}
