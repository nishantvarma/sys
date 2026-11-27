// entries are dir's children as fm shows them: first the folder itself
// and its parent, . and .., which ReadDir never lists, whatever else
// shows, for they are the folder's own handles; then its children by rank
// and by name, dotfiles only when hidden.
func entries(dir string, hidden bool) []entry {
	var files []entry
	if ents, err := os.ReadDir(dir); err == nil {
		for _, e := range ents {
			if hidden || !strings.HasPrefix(e.Name(), ".") {
				files = append(files, entryOf(dir, e))
			}
		}
		slices.SortStableFunc(files, func(a, b entry) int {
			if r := rank(a) - rank(b); r != 0 {
				return r
			}
			an := strings.ToLower(a.name)
			bn := strings.ToLower(b.name)
			return strings.Compare(an, bn)
		})
	}
	return append([]entry{
		{name: ".", path: dir, dir: true},
		{name: "..", path: filepath.Dir(dir), dir: true},
	}, files...)
}
