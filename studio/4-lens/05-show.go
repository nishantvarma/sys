// showVerb is the lens as a program: the thing at cur through the
// window's lens, as nodes that a view draws, as fm's are. On the terminal
// its view is sys/view's, on the whole of its pane; with -w it is Tk's,
// and with -s a host's at the other end of stdin and stdout, which starts
// it without the eye, as fm -s is. Outside tmux its window is none, as
// fm's is.
func showVerb(args []string) error {
	mode := ""
	if len(args) > 0 {
		mode = args[0]
	}
	if len(args) > 1 || mode != "" && mode != "-w" && mode != "-s" {
		return errors.New("usage: studio show [-w | -s]")
	}
	win, eye := state.Pane()
	l := &lensing{curf: state.Path("cur", win),
		lensf: state.Path("lens", win), on: eye && mode != "-s"}
	for _, p := range []string{l.curf, l.lensf} {
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			return err
		}
	}
	wt, err := watch.New(filepath.Dir(l.curf), filepath.Dir(l.lensf))
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
	defer func() {
		t.Write(tty.FocusOff + tty.AltOff + tty.Cnorm)
		t.Flush()
		t.Close()
	}()
	home, _ := os.UserHomeDir()
	v := view.New(t, home, looks)
	v.Chips([2]string{"m", "next lens"})
	return l.run(v, t, wt)
}
