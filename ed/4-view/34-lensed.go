// lensed is the screen, w by h, while a lens other than source shows: the
// bar, then sys/view in the region under it, drawing the text's items
// (itemsOf) as fm draws a folder, with ed's word and state at the bottom.
// The view writes nothing itself: its lines come back (view.Lines), for
// tty's Frame to write with the bar.
func (v *term) lensed(
	n *outline.Node,
	status string,
	state []string,
	w, h int,
) []string {
	if v.seen == nil {
		home, _ := os.UserHomeDir()
		v.seen = view.New(nil, home, nil)
	}
	top := tops(h)
	v.seen.In(0, top, w, h-top)
	v.items = v.itemsOf(n)
	v.seen.Draw(v.items, status, state)
	var rows []string
	if top > 0 {
		bar, _ := v.bar(w)
		rows = append(rows, bar)
	}
	for _, l := range v.seen.Lines() {
		rows = append(rows, view.Paint(l))
	}
	return rows
}
