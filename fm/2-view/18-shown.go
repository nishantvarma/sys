// shown is what the view shows of the tree drawn last: as the folder or as
// tiles, the cursor's folder and its rows; as the tree, every open row from
// the top, under the top's path.
func (v *term) shown() (*outline.Node, []row) {
	if v.lay == "tree" {
		return v.last, flat(v.last, 0, "")
	}
	z := zoom(v.last)
	if z == nil {
		z = v.last
	}
	rows := make([]row, len(z.Kids))
	for i, k := range z.Kids {
		rows[i] = row{n: k}
	}
	return z, rows
}
