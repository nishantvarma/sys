// follow keeps the selection on its text when line l gains (d = 1) or
// loses (d = -1) a rune at its start, as indent and dedent make it. An
// end past column 0 moves with its rune; an end at column 0 holds the
// line start, so whole lines stay whole. No clamp: each end moves
// exactly as far as its line did.
func (e *editor) follow(l, d int) {
	for _, p := range []*pos{&e.caret.anchor, &e.caret.head} {
		if p.line == l && p.col > 0 {
			p.col += d
		}
	}
	e.goal = e.caret.head.col
}
