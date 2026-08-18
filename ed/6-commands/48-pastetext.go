// pasteText drops a terminal paste in at the caret, replacing a selection
// as p does. The terminal brackets it, so the whole block arrives as one
// key and costs one snapshot and one redraw — not one of each per rune.
// Control runes are stripped: a paste is text, and its keystroke reading
// was only ever an accident of how it used to arrive.
func (e *editor) pasteText(s string) {
	rs := slices.DeleteFunc([]rune(s), func(r rune) bool {
		return r != '\n' && r != '\t' && !unicode.IsPrint(r)
	})
	if len(rs) == 0 {
		return
	}
	e.b.snapshot(e.caret)
	a, c := e.rng()
	if a != c {
		e.b.cut(a, c)
	}
	p := e.b.paste(e.b.clamp(a, true), rs)
	if !e.mode.pastEnd() { // normal mode sits on the last rune
		p = back(e.b, p)
	}
	e.place(p)
}
