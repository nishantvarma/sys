// reorder moves the row under the cursor d places, -1 up and 1 down, when
// it and the row it passes are numbered (number): the two swap their
// numbers, and the cursor goes with its row. A name either would take that
// any other entry holds, a broken link too, stops it.
func (m *fm) reorder(d int) {
	i, j := m.idx, m.idx+d
	if j < 0 || j >= len(m.files) {
		return
	}
	a, b := m.files[i], m.files[j]
	na, nb := number(a.name), number(b.name)
	switch {
	case na == "":
		m.flash("not numbered: " + a.name)
		return
	case nb == "":
		m.flash("not numbered: " + b.name)
		return
	case na == nb:
		m.flash("the same number: " + a.name + " " + b.name)
		return
	}
	toA := filepath.Join(m.cwd, nb+a.name[len(na):])
	toB := filepath.Join(m.cwd, na+b.name[len(nb):])
	for _, p := range []string{toA, toB} {
		if _, err := os.Lstat(p); err == nil && p != a.path &&
			p != b.path {
			m.flash("taken: " + filepath.Base(p))
			return
		}
	}
	// through a third name, for the two may swap the very same names; a
	// rename that fails puts back what moved, and says where a file stays
	// when that fails too
	tmp := filepath.Join(m.cwd, ".reorder."+strconv.Itoa(os.Getpid()))
	back := func(from, to string) {
		if os.Rename(from, to) != nil {
			m.flash("reorder failed: " + filepath.Base(to) +
				" is kept as " + filepath.Base(from))
		}
	}
	switch {
	case !m.catch(os.Rename(a.path, tmp)):
		return
	case !m.catch(os.Rename(b.path, toB)):
		back(tmp, a.path)
		return
	case !m.catch(os.Rename(tmp, toA)):
		back(toB, b.path)
		back(tmp, a.path)
		return
	}
	delete(m.sel, a.path)
	delete(m.sel, b.path)
	m.ls()
	m.land(filepath.Base(toA))
}
