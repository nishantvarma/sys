// pick runs a finder and lets fzy choose one of its lines. The list is read
// whole: piped, a finder outrunning a dead fzy blocks on the end we hold.
func (m *fm) pick(args ...string) string {
	m.t.Suspend()
	defer m.t.Resume()
	tookInterrupt() // stale Ctrl-C; the children own the next one
	find := exec.Command(args[0], args[1:]...)
	find.Dir, find.Stderr = m.ctx(), os.Stderr
	list, _ := find.Output() // no hits is an exit status, not an error
	if len(list) == 0 {
		return ""
	}
	fzy := exec.Command("fzy")
	fzy.Dir, fzy.Stderr = m.ctx(), os.Stderr
	fzy.Stdin = bytes.NewReader(list)
	out, err := fzy.Output()
	if err != nil || tookInterrupt() {
		return ""
	}
	return strings.TrimSpace(string(out))
}
