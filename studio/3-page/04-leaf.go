// leaf writes a slot that holds a pane: its name, what runs there, the
// tabs that take turns in it when it has more than one, and what its page
// view waits on. None is built yet, so each is a placeholder that says so.
func leaf(b *strings.Builder, s *slot) {
	cmd := s.cmd
	if cmd == "" {
		cmd = "a shell"
	}
	fmt.Fprintf(b, "<div class=\"pane\" id=\"slot-%s\">\n"+
		"<p class=\"name\">%s <code>%s</code></p>\n",
		html.EscapeString(s.name), html.EscapeString(s.name),
		html.EscapeString(cmd))
	if len(s.tabs) > 1 {
		var names []string
		for _, tb := range s.tabs {
			names = append(names, html.EscapeString(tb.name))
		}
		fmt.Fprintf(b, "<p class=\"tabs\">tabs: %s</p>\n",
			strings.Join(names, " · "))
	}
	fmt.Fprintf(b, "<p class=\"waits\">%s</p>\n</div>\n",
		html.EscapeString(waits(s.cmd)))
}
