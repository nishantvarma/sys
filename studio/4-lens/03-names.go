// names is the lenses' names, in order, as a name it does not know is
// told them (show).
func names() []string {
	var out []string
	for _, l := range lenses {
		out = append(out, l.name)
	}
	return out
}
