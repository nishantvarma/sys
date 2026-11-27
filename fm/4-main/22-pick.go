// pick runs a finder and lets fzy choose one of its lines, on fm's
// terminal as spawn runs a child: the one fm is drawn on, suspended, or
// its console, where fzy finds its terminal, /dev/tty, as it finds any.
// The list is read whole: piped, a finder outrunning a dead fzy blocks on
// the end we hold. With no terminal, in a window, fzy has none to run on,
// and the pick says so, as what else needs the terminal does (spawn).
func (m *fm) pick(args ...string) string {
	if m.t == nil && m.cons == nil {
		m.flash("pick: needs the terminal")
		return ""
	}
	if m.t != nil {
		m.t.Suspend()
		defer m.t.Resume()
	}
	tookInterrupt() // stale Ctrl-C; the children own the next one
	find := exec.Command(args[0], args[1:]...)
	find.Dir, find.Stderr = m.cwd, os.Stderr
	list, _ := find.Output() // no hits is an exit status, not an error
	if len(list) == 0 {
		return ""
	}
	fzy := exec.Command("fzy")
	fzy.Dir, fzy.Stderr = m.cwd, os.Stderr
	fzy.Stdin = bytes.NewReader(list)
	out, err := fzy.Output()
	if err != nil || tookInterrupt() {
		return ""
	}
	return strings.TrimSpace(string(out))
}
