// picturePage is the skeleton with a note's picture in it: its title, its
// lead, the picture, and the note it was drawn from, folded under it: what
// the lens's render shows (`design/lens`), on a page. Its look is the
// picture's own (picture.Style), under local (fill).
func picturePage(f *picture.Picture, skel, note, src string) string {
	var m strings.Builder
	fmt.Fprintf(&m, "<main>\n    <h1>%s</h1>\n",
		html.EscapeString(f.Title))
	if f.Lead != "" {
		fmt.Fprintf(&m, "    <div class=\"lead\">\n      <p>%s</p>\n"+
			"    </div>\n", html.EscapeString(f.Lead))
	}
	fmt.Fprintf(&m, pictureNodes, f.SVG(), html.EscapeString(note),
		html.EscapeString(note), html.EscapeString(src))
	return fill(skel, f.Title, picture.Style, m.String())
}
