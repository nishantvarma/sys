// curRow is the cursor's row: the one it is on, else the folder it is in,
// as an empty folder holds it, else the first.
func curRow(rows []row) int {
	for i, r := range rows {
		if r.n.Cur {
			return i
		}
	}
	for i, r := range rows {
		if r.n.In {
			return i
		}
	}
	return 0
}
