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
