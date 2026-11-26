// lensed is the screen, w by h, while a lens other than source shows:
// sys/view drawing the text's items (itemsOf) as fm draws a folder, with
// ed's word and state at the bottom, and the bar in place of the view's
// head, so that it stands over the view's rule as over source's. The
// view writes nothing itself: its lines come back (view.Lines), for tty's
// Frame to write.
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
	v.seen.In(0, 0, w, h)
	v.items = v.itemsOf(n)
	v.seen.Draw(v.items, status, state)
	var rows []string
	for _, l := range v.seen.Lines() {
		rows = append(rows, view.Paint(l))
	}
	if tops(h) > 0 && len(rows) > 0 {
		rows[0], _ = v.bar(w)
	}
	return rows
}
