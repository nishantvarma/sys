// run draws the lens and takes its keys until it is quit, outside a frame,
// for in one q is fm's, or its view is gone. A change to cur moves it
// there; a key on a node reaches it first (see reach), and one on a node
// gone goes with it; m turns to the next lens. Its drawing is a tree, the
// program's word what made it, and its state the lens and its number
// among them; a view that has tabs shows the lenses as tabs (tabs).
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
		ln, i, ok := lensNamed(state.Read(l.lensf))
		if !ok {
			ln = lenses[0]
		}
		path, _ := outline.Address(l.at)
		if fi, err := os.Stat(path); err == nil {
			wt.Follow(folder(path, fi))
		}
		word := ln.by
		if l.msg != "" {
			word, l.msg = l.msg, ""
		}
		n := fmt.Sprintf("%s %d/%d", ln.name, i+1, len(lenses))
		t := tree(l.at, ln)
		if tv, ok := ui.(interface{ Tabs([]view.Tab) }); ok {
			tv.Tabs(tabs(i, t))
		}
		ui.Draw(t, word, []string{n})
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
		case "m":
			next := lenses[(i+1)%len(lenses)].name
			if err := state.Write(l.lensf, next); err != nil {
				l.msg = err.Error()
			}
		}
	}
}
