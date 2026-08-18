func (m *fm) ls() {
	name := m.curName()
	m.files = nil
	if m.hidden { // ls skipped these two by their dot, and hid the rest
		m.files = []entry{
			{name: ".", path: m.cwd, dir: true},
			{name: "..", path: filepath.Dir(m.cwd), dir: true},
		}
	}
	if ents, err := os.ReadDir(m.cwd); err == nil {
		for _, e := range ents {
			if m.hidden || !strings.HasPrefix(e.Name(), ".") {
				m.files = append(m.files, entryOf(m.cwd, e))
			}
		}
		slices.SortStableFunc(m.files, func(a, b entry) int {
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
	// the cursor keeps its name across the rebuild, else its row
	if i := m.at(name); i >= 0 {
		m.idx = i
	} else {
		m.idx = max(0, min(m.idx, len(m.files)-1))
	}
}
