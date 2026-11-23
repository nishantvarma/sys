// page is the skeleton with the frame in it: its title, its lead, its
// panes as the terminal lays them out, and the note it was read from,
// with the frame's look under local (fill).
func (f *frame) page(skel string) string {
	var b strings.Builder
	boxes(&b, f.slots, false)
	name := f.note
	if r, err := filepath.Rel(root(), f.note); err == nil &&
		!strings.HasPrefix(r, "..") {
		name = r
	}
	m := fmt.Sprintf(body, html.EscapeString(f.title),
		html.EscapeString(f.lead), b.String(), html.EscapeString(name),
		html.EscapeString(f.src))
	return fill(skel, f.title, local, m)
}
