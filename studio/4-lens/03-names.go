// names is the lenses' names, in order, which the lens puts on its tree's
// tags for a view's bar (`view/bar.go:/func Bar/`).
func names() []string {
	var out []string
	for _, l := range lenses {
		out = append(out, l.name)
	}
	return out
}
