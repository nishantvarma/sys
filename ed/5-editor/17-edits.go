// edits is what ed -f edits for an address cur holds: a file, from its
// line; a folder's README, the folder's own note, from its top. Nothing
// for the rest: a folder with no README, a file gone, or one that is no
// note a person types (outline.File).
func edits(at string) (string, int, bool) {
	path, line := outline.Address(at)
	fi, err := os.Stat(path)
	if err != nil {
		return "", 0, false
	}
	if fi.IsDir() {
		notes := []string{"README.md", "README", "readme.md"}
		for _, r := range notes {
			p := filepath.Join(path, r)
			if _, err := outline.File(p); err == nil {
				return p, 0, true
			}
		}
		return "", 0, false
	}
	_, err = outline.File(path)
	return path, line, err == nil
}
