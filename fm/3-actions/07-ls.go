func (m *fm) ls() {
	m.files = nil
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
	m.idx = tty.Clamp(m.idx, 0, max(0, len(m.files)-1))
}
