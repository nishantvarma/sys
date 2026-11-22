// toward is the name of dir's row on the way down to path: empty when path
// is dir itself or not below it.
func toward(dir, path string) string {
	if dir == path || !under(path, dir) {
		return ""
	}
	rest := strings.TrimPrefix(path, strings.TrimSuffix(dir, "/")+"/")
	name, _, _ := strings.Cut(rest, "/")
	return name
}
