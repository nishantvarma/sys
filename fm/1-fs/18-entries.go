// entries are dir's children as fm shows them: folders and links first,
// then by name; dotfiles only when hidden, and with dots, . and .. too.
func entries(dir string, hidden, dots bool) []entry {
	var files []entry
	if hidden && dots { // ReadDir never lists these two
		files = []entry{
			{name: ".", path: dir, dir: true},
			{name: "..", path: filepath.Dir(dir), dir: true},
		}
	}
	if ents, err := os.ReadDir(dir); err == nil {
		for _, e := range ents {
			if hidden || !strings.HasPrefix(e.Name(), ".") {
				files = append(files, entryOf(dir, e))
			}
		}
		slices.SortStableFunc(files, func(a, b entry) int {
			ai, bi := a.dir || a.link, b.dir || b.link
			if ai != bi { // dirs and symlinks first
				if ai {
					return -1
				}
				return 1
			}
			an := strings.ToLower(a.name)
			bn := strings.ToLower(b.name)
			return strings.Compare(an, bn)
		})
	}
	return files
}
