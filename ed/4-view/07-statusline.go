// statusLine is the bottom line, row h, w wide: the mode in its colour,
// bold, then the file, + while unsaved and held while that keeps ed from
// cur's file (chase), or in their place a passing message; line:col at
// the right.
func statusLine(
	n *outline.Node,
	state []string,
	status string,
	line, col, w, h int,
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
	body := tty.Paint(" "+mode, modeColour(mode), tty.Ground, true) +
		tty.Paint("  "+info, fg, tty.Ground, false) +
		tty.Paint(strings.Repeat(" ", max(0, w-used)), "", tty.Ground,
			false) +
		tty.Paint(right, tty.Soft, tty.Ground, false)
	return fmt.Sprintf("\x1b[%d;1H%s%s", h, tty.ClrEol, body)
}
