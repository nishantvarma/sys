// statusLine: colour-coded mode + filename on the left, line:col on the right.
// A transient message takes the filename's place.
func statusLine(
	n *outline.Node,
	state []string,
	status string,
	line, col, w, h int,
) string {
	mode := state[0]
	info := tilde(n.Text)
	if slices.Contains(state[1:], "+") {
		info += " +"
	}
	if status != "" {
		info = status
	}
	right := fmt.Sprintf("%d:%d", line+1, col+1)
	plain := mode + "  " + info
	if keep := w - len([]rune(right)) - 1; len([]rune(plain)) > keep {
		if keep < 0 {
			keep = 0
		}
		plain = string([]rune(plain)[:keep])
	}
	pad := w - len([]rune(plain)) - len([]rune(right))
	if pad < 1 {
		pad = 1
	}
	body := plain + strings.Repeat(" ", pad) + right
	// Colour the mode word in place; escapes carry no width, so
	// alignment holds.
	if strings.HasPrefix(body, mode) {
		body = modeColour(mode) + body[len(mode):]
	}
	return fmt.Sprintf("\x1b[%d;1H%s%s", h, tty.ClrEol, body)
}
