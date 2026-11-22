// file is the text file the cursor's item is in: the item's id short of
// its line; empty on a row.
func (m *fm) file() string {
	return m.item[:max(0, strings.LastIndex(m.item, ":"))]
}
