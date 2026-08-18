// statusLine: colour-coded mode + filename on the left, line:col on the right.
// A transient message takes the filename's place.
func statusLine(f frame) string {
	info := tilde(f.path)
	if f.dirty {
		info += " +"
	}
	if f.msg != "" {
		info = f.msg
	}
	right := fmt.Sprintf("%d:%d", f.head.line+1, f.head.col+1)
	plain := f.mode + "  " + info
	if keep := f.w - len([]rune(right)) - 1; len([]rune(plain)) > keep {
		if keep < 0 {
			keep = 0
		}
		plain = string([]rune(plain)[:keep])
	}
	pad := f.w - len([]rune(plain)) - len([]rune(right))
	if pad < 1 {
		pad = 1
	}
	body := plain + strings.Repeat(" ", pad) + right
	// Colour the mode word in place; escapes carry no width, so
	// alignment holds.
	if strings.HasPrefix(body, f.mode) {
		body = modeColour(f.mode) + body[len(f.mode):]
	}
	return fmt.Sprintf("\x1b[%d;1H%s%s", f.h, tty.ClrEol, body)
}
