// statusLine is the bottom line, w wide: the mode in its colour, bold,
// then the file, + while unsaved and held while that keeps ed from cur's
// file (chase), or in their place a passing message; line:col at the
// right.
func statusLine(
	n *outline.Node,
	state []string,
	status string,
	line, col, w int,
) string {
	mode := state[0]
	info, fg := tilde(n.Text), tty.Soft
	if slices.Contains(state[1:], "+") {
		info += " +"
	}
	if slices.Contains(state[1:], "held") {
		info += " held"
	}
	if status != "" {
		info, fg = status, tty.Fore
	}
	right := fmt.Sprintf("%d:%d ", line+1, col+1)
	room := max(0, w-tty.Width(right)-1-tty.Width(" "+mode+"  "))
	if r := []rune(info); len(r) > room {
		info = string(r[:room])
	}
	used := tty.Width(" " + mode + "  " + info + right)
	return tty.Paint(" "+mode, modeColour(mode), tty.Ground, true) +
		tty.Paint("  "+info, fg, tty.Ground, false) + blank(w-used) +
		tty.Paint(right, tty.Soft, tty.Ground, false)
}
