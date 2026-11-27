// hidden is the tab a hidden panel shows again, of its tabs tps: the one
// it hid, @studio-hidden (hide), else its first.
func (t *tmux) hidden(win string, tps []tabPane) tabPane {
	was, _ := t.run("show", "-wqv", "-t", win, "@studio-hidden")
	for _, tp := range tps {
		if tp.pane == was {
			return tp
		}
	}
	return tps[0]
}
