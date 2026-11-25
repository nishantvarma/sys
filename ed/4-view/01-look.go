// tabWidth is how far a tab runs, to the next stop.
const tabWidth = 4

// gutter is the columns before a line's text, for a buffer of n lines: a
// margin, the line's number in three digits or as many as the last
// needs, and two blanks.
func gutter(n int) int {
	return max(3, len(strconv.Itoa(n))) + 3
}
