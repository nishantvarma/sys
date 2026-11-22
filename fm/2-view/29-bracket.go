// bracket draws the left side of a box down the rows of a file's items, in
// the indent that starts at byte at of each lead: ╭ on the first row, │ on
// those between and ╰ on the last, so a file's insides read apart from the
// folders around them.
func bracket(rows []row, at int) {
	for i := range rows {
		g := "│ "
		switch {
		case i == len(rows)-1:
			g = "╰ "
		case i == 0:
			g = "╭ "
		}
		l := rows[i].lead
		rows[i].lead = l[:at] + g + l[at+len("  "):]
	}
}
