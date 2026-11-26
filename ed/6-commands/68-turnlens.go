// turnLens shows the next lens, and source again after the last; toLens
// the one named, as a press on its name on a view's bar says. Each view
// of ed draws what ed says, so all of them show the lens ed shows.
func (e *editor) turnLens() {
	e.lens = (e.lens + 1) % len(lenses)
}

func (e *editor) toLens(name string) {
	if i := slices.Index(lenses, name); i >= 0 {
		e.lens = i
	}
}
