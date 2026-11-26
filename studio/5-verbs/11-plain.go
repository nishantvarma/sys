// plain makes a frame's window a plain one again: its hooks, its borders,
// the options studio set on it and the names its panes carry gone (open),
// and @studio too, when it names this window.
func plain(t *tmux, win string) {
	for _, h := range hooks("") {
		t.run("set-hook", "-wu", "-t", win, h[0])
	}
	for _, o := range borders {
		t.run("set", "-wu", "-t", win, o)
	}
	panes, _ := t.run("list-panes", "-t", win, "-F", "#{pane_id}")
	for _, p := range strings.Fields(panes) {
		t.run("set", "-pu", "-t", p, "@slot")
	}
	opts, _ := t.run("show", "-w", "-t", win)
	for _, l := range strings.Split(opts, "\n") {
		if name, _, _ := strings.Cut(l, " "); strings.HasPrefix(name,
			"@studio") {
			t.run("set", "-wu", "-t", win, name)
		}
	}
	if g, _ := t.run("show", "-gqv", "@studio"); g == win {
		t.run("set", "-gu", "@studio")
	}
}
