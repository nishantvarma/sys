// pick runs a finder and lets fzy choose one of its lines. The list is read
// whole: piped, a finder outrunning a dead fzy blocks on the end we hold.
// Off the terminal, in a window, fzy runs beside it (beside).
func (m *fm) pick(args ...string) string {
	if m.t == nil {
		find := exec.Command(args[0], args[1:]...)
		find.Dir = m.cwd
		list, _ := find.Output()
		return m.beside(list)
	}
	m.t.Suspend()
	defer m.t.Resume()
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

// beside is a pick off the terminal: fzy, which needs one, runs on the
// list in a terminal beside the window, st, as `sh/spawn` opens one, and
// its choice comes back in a file, once it closes. fm waits for it, as it
// waits for fzy on its own terminal.
func (m *fm) beside(list []byte) string {
	if len(list) == 0 {
		return ""
	}
	dir, err := os.MkdirTemp("", "fm-pick")
	if err != nil {
		m.flash("pick: " + err.Error())
		return ""
	}
	defer os.RemoveAll(dir)
	in, out := filepath.Join(dir, "list"), filepath.Join(dir, "pick")
	if err := os.WriteFile(in, list, 0o600); err != nil {
		m.flash("pick: " + err.Error())
		return ""
	}
	st := exec.Command("st", "-e", "sh", "-c", `fzy <"$1" >"$2"`, "pick",
		in, out)
	st.Dir = m.cwd
	if err := st.Run(); err != nil {
		m.flash("pick: st: " + err.Error())
		return ""
	}
	b, _ := os.ReadFile(out)
	return strings.TrimSpace(string(b))
}
