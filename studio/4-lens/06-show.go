// showVerb is the lens as a program: the thing at cur seen one way, the
// lens named, a tab of the frame's lens view (`design/lens`), or, with
// none, every way, its own Tab turning them, as nodes that a view draws, as
// fm's are. On the terminal its view is sys/view's, on the whole of its
// pane; with -w the window's, and with -s a host's at the other end of
// stdin and stdout, which starts it without the eye, as fm -s is. Outside
// tmux its window is none, as fm's is. It follows cur and, for a lens that
// reads the text, what ed in the frame holds unsaved (noted), and so draws
// again as either changes.
func showVerb(args []string) error {
	l := &lensing{}
	mode := ""
	for _, a := range args {
		switch {
		case a == "-w" || a == "-s":
			mode = a
		case !l.fixed:
			_, i, ok := lensNamed(a)
			if !ok {
				return fmt.Errorf("%s: no such lens; there"+
					" are %s", a,
					strings.Join(names(), ", "))
			}
			l.lens, l.fixed = i, true
		default:
			return errors.New("usage: studio show [lens]" +
				" [-w | -s]")
		}
	}
	win, eye := state.Pane()
	l.curf = state.Path("cur", win)
	l.unsavedf = state.Path("unsaved", win)
	l.textf = state.Path("text", win)
	l.slot = state.Slot()
	l.on, l.framed = eye && mode != "-s", l.slot != ""
	follows := []string{l.curf}
	if !l.fixed || lenses[l.lens].typed {
		follows = append(follows, l.unsavedf, l.textf)
	}
	var dirs []string
	for _, p := range follows {
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			return err
		}
		dirs = append(dirs, filepath.Dir(p))
	}
	wt, err := watch.New(dirs...)
	if err != nil {
		return err
	}
	switch mode {
	case "-w":
		ui, done, err := outline.Window()
		if err != nil {
			return err
		}
		defer done()
		return l.run(ui, nil, wt)
	case "-s":
		return l.run(outline.Stream(os.Stdin, os.Stdout), nil, wt)
	}
	t, err := tty.New()
	if err != nil {
		return err
	}
	t.Write(tty.AltOn + tty.Civis)
	t.Flush()
	t.Focus()
	t.Mouse()
	defer func() {
		t.Write(tty.MouseOff + tty.FocusOff + tty.AltOff + tty.Cnorm)
		t.Flush()
		t.Close()
	}()
	home, _ := os.UserHomeDir()
	return l.run(view.New(t, home), t, wt)
}
