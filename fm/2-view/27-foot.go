// foot is the bottom line, w wide: the program's word on the left, and on
// the right its state, then the view's, dim. Short of room, the word is
// cut, for the state says what holds now.
func foot(word string, state, mine []string, w int) string {
	right := strings.Join(mine, " ")
	if len(state) > 0 {
		right = strings.Join(state, " ") + " · " + right
	}
	right = tty.Clip(right, w)
	left := tty.Clip(word, max(0, w-tty.Width(right)-1))
	pad := max(0, w-tty.Width(left)-tty.Width(right))
	return left + strings.Repeat(" ", pad) + tty.Dim(right)
}
