// drag moves the selection's head to p, where the mouse drags, its anchor
// staying where the press put the cursor, both on runes: a drag past a
// line's end stops at its last, as in any editor. The first move of a drag
// goes into insert, for a selection made with the mouse is typed over;
// esc leaves it selected, for the verbs.
func (e *editor) drag(p pos) {
	if e.mode != e.ins {
		e.mode, e.owed = e.ins, true
	}
	e.unkept = true
	e.caret.anchor = e.b.clamp(e.caret.anchor, false)
	e.caret.head = e.b.clamp(p, false)
	e.goal = e.caret.head.col
}
