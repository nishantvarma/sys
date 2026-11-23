// leaf writes a slot that holds a pane: its name, what runs there, and
// what its page view waits on. None is built yet, so each is a
// placeholder that says so.
func leaf(b *strings.Builder, s *slot) {
	cmd := s.cmd
	if cmd == "" {
		cmd = "a shell"
	}
	fmt.Fprintf(b, "<div class=\"pane\" id=\"slot-%s\">\n"+
		"<p class=\"name\">%s <code>%s</code></p>\n"+
		"<p class=\"waits\">%s</p>\n</div>\n",
		html.EscapeString(s.name), html.EscapeString(s.name),
		html.EscapeString(cmd), html.EscapeString(waits(s.cmd)))
}
