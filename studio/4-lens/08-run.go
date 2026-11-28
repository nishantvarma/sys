// run draws the lens and takes its keys until it is quit, outside a frame,
// for in one q is fm's, or its view is gone. A change to cur moves it
// there; a key on a node reaches it first (see reach), and one on a node
// gone goes with it; enter, a right click's too, opens what the row names
// (plumb), and in a doc the word it fell on. Tab turns to the next lens:
// for a lens fixed, a tab of the frame's lens view, the frame's next tab
// (turn); else its own next. Its drawing is a tree, the program's word how
// the lens sees the thing, and its state the lens shown and a tally of
// what it shows.
func (l *lensing) run(ui outline.View, t *tty.Term,
	wt *watch.Watcher) error {
	var keys chan keyed
	if t != nil {
		go func() {
			for range wt.On {
				t.Wake()
			}
		}()
	} else {
		keys = make(chan keyed)
		go func() {
			for {
				ev, ok := ui.Key()
				keys <- keyed{ev, ok}
				if !ok {
					return
				}
			}
		}()
	}
	for {
		if c := state.Read(l.curf); c != "" && c != l.said {
			l.said, l.at = c, c
		}
		if l.at == "" {
			l.at, _ = os.Getwd() // with none, the lens says why
		}
		ln := lenses[l.lens]
		th := l.see(l.at)
		if fi, err := os.Stat(th.path); err == nil {
			wt.Follow(folder(th.path, fi))
		}
		tr, word := tree(th, ln)
		switch {
		case th.held && ln.typed:
			word += ", as ed holds it unsaved"
		case th.held:
			word += ", of the file as saved"
		}
		if l.msg != "" {
			word, l.msg = l.msg, ""
		}
		shows := []string{ln.name}
		if n := tallyOf(tr); n != "" {
			shows = append(shows, n)
		}
		ui.Draw(tr, word, shows)
		ev, ok := take(ui, keys, wt)
		if !ok {
			return nil
		}
		if ev.Node != "" && !l.reach(ev.Node) {
			l.msg = "gone: " + ev.Node
			continue
		}
		switch ev.Name {
		case "q":
			if !l.framed {
				return nil
			}
			l.msg = "q in fm ends the frame"
		case tty.KFocus, tty.KBlur:
			l.on = ev.Name == tty.KFocus
		case tty.KTab:
			switch {
			case !l.fixed:
				l.lens = (l.lens + 1) % len(lenses)
			case !l.framed:
				l.msg = "no tabs outside a frame"
			default:
				err := turn(l.slot, os.Getenv("TMUX_PANE"),
					[]string{"next"})
				if err != nil {
					l.msg = err.Error()
				}
			}
		case tty.KEnter:
			to := l.at // or, in a doc, the word a click fell on
			if ev.Text != "" {
				to = ev.Text
			}
			plumb(to)
		}
	}
}
