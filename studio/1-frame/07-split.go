// split cuts an item at its first two spaces: the name, and the rest.
func split(text string) (name, rest string) {
	name, rest, _ = strings.Cut(text, "  ")
	return strings.TrimSpace(name), strings.TrimSpace(rest)
}
