// rank is where an entry stands among its folder's: . and .. first, then
// what is numbered (number), folders among files, then folders and links,
// then files.
func rank(f entry) int {
	switch {
	case f.name == "." || f.name == "..":
		return 0
	case number(f.name) != "":
		return 1
	case f.dir || f.link:
		return 2
	}
	return 3
}
