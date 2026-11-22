// inside fills k, the row of path p, with what opening it shows, by what p
// is: a folder, or a link to one, shows its rows; a text file its items,
// read by their indent. Anything else stays shut. It is the format
// dispatcher in place, as o is the one that opens a file beside fm.
func (m *fm) inside(k *outline.Node, p string) {
	if isDir(p) {
		sub := m.tree(p, k.Text)
		k.Open, k.In = true, sub.In
		k.Kids = sub.Kids
	} else if kids, ok := read(p); ok {
		k.Open, k.Kids = true, kids
		m.flag(k, p)
	}
}
