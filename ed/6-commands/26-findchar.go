// findChar reads one more key and selects from the cursor to (till) or
// through (!till) its next occurrence on the current line, forward or
// backward. No match, or a non-rune key, leaves the selection untouched.
// A set ext keeps the standing anchor.
func (e *editor) findChar(forward, till bool) {
	e.msg = "find:"
	e.draw()
	k, ok := e.t.ReadKey()
	e.msg = ""
	if !ok || k.Name != "" {
		return
	}
	from := e.head()
	found, hit := findOnLine(e.b, from, k.Rune, forward)
	if !hit {
		return
	}
	if till {
		if forward {
			found.col--
		} else {
			found.col++
		}
	}
	if !e.ext {
		e.caret.anchor = from
	}
	e.caret.head = e.b.clamp(found, false)
	e.goal = e.caret.head.col
}
