// kind is what an entry is, the word a view picks its look by.
func kind(f entry) string {
	switch {
	case f.bad:
		return "bad"
	case f.link:
		return "link"
	case f.dir:
		return "dir"
	case f.exec:
		return "exec"
	default:
		return "file"
	}
}
