const countMax = 9999 // caps a fat-fingered prefix so the replay can't hang

func (normalMode) key(e *editor, k tty.Key) {
	tok := k.Token()
	if tok != "q" { // any other key disarms a pending dirty-quit
		e.armed = false
	}
	// A leading digit builds a repeat count; a bare 0 stays the line-start
	// motion, so 0 only joins a count already under way.
	if len(tok) == 1 && '0' <= tok[0] && tok[0] <= '9' &&
		(tok != "0" || e.count > 0) {
		e.count = min(e.count*10+int(tok[0]-'0'), countMax)
		return
	}
	act := normalMap[tok]
	if act == nil {
		e.count = 0
		return
	}
	// A count just replays the keystroke: 3w is w w w. The first press
	// runs here; the rest re-enter through the current mode, so a key
	// that switched to insert (i, c, o) lets insert absorb them.
	n := max(1, e.count)
	e.count = 0
	act(e)
	for i := 1; i < n; i++ {
		e.mode.key(e, k)
	}
}
