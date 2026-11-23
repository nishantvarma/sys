// tabs are the lenses as the view's switch line shows them: each by its
// name, the one in use, on, lit with a tally of what it shows, t
// (tallyOf), and a placeholder dim, for it waits to be built.
func tabs(on int, t *outline.Node) []view.Tab {
	var out []view.Tab
	for i, l := range lenses {
		tab := view.Tab{Name: l.name, On: i == on,
			Waits: l.nodes == nil}
		if i == on {
			tab.Tally = tallyOf(t)
		}
		out = append(out, tab)
	}
	return out
}
