// figurePage is the skeleton with a note's figure in it: its title, its
// lead, the figure, and the note it was drawn from, folded under it; the
// figure lens's page, until the page hosts the lens itself. Its look is
// the figure's own (figure.Style), under local (fill).
func figurePage(f *figure.Figure, skel, note, src string) string {
	var m strings.Builder
	fmt.Fprintf(&m, "<main>\n    <h1>%s</h1>\n",
		html.EscapeString(f.Title))
	if f.Lead != "" {
		fmt.Fprintf(&m, "    <div class=\"lead\">\n      <p>%s</p>\n"+
			"    </div>\n", html.EscapeString(f.Lead))
	}
	fmt.Fprintf(&m, figureNodes, f.SVG(), html.EscapeString(note),
		html.EscapeString(note), html.EscapeString(src))
	return fill(skel, f.Title, figure.Style, m.String())
}
