// lensVerb says the window's lens, or sets it by its name or number. A
// lens is state, so it is a file, as cur is: any program sets it, and the
// lens, watching, draws again.
func lensVerb(args []string) error {
	if len(args) > 1 {
		return errors.New("usage: studio lens [name]")
	}
	win, err := target()
	if err != nil {
		return err
	}
	p := state.Path("lens", win)
	if len(args) == 0 {
		l, _, ok := lensNamed(state.Read(p))
		if !ok {
			l = lenses[0]
		}
		fmt.Println(l.name)
		return nil
	}
	l, _, ok := lensNamed(args[0])
	if !ok {
		var names []string
		for _, each := range lenses {
			names = append(names, each.name)
		}
		return fmt.Errorf("%s: no such lens; there are %s", args[0],
			strings.Join(names, ", "))
	}
	return state.Write(p, l.name)
}
