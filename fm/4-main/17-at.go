// at is the row named name, or -1.
func (m *fm) at(name string) int {
	for i, f := range m.files {
		if f.name == name {
			return i
		}
	}
	return -1
}
