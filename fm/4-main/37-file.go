// file is the text file the cursor's item is in: the item's address short
// of its line; empty on a row.
func (m *fm) file() string {
	p, _ := outline.Address(m.item)
	return p
}
