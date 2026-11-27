// rank is where an entry stands among its folder's children: what is
// numbered (number), folders among files, first; then folders and links;
// then files.
func rank(f entry) int {
	switch {
	case number(f.name) != "":
		return 0
	case f.dir || f.link:
		return 1
	}
	return 2
}
