// drag moves the selection's head to p, where the mouse drags, its anchor
// staying where the press put the cursor, both on runes: a drag past a
// line's end stops at its last, as in any editor. It selects as a motion
// does, in the mode it finds: in normal the selection is the verbs', and
// insert is only i's, or a click on the mode's word. A drag in an insert
// makes that insert one . cannot type again, for the mouse is no key.
func (e *editor) drag(p pos) {
	e.unkept = e.unkept || e.mode == e.ins
	e.caret.anchor = e.b.clamp(e.caret.anchor, false)
	e.caret.head = e.b.clamp(p, false)
	e.goal = e.caret.head.col
}
