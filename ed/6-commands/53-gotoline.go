// goToLine reads a 1-based line number at the prompt and places the
// cursor at its start; junk input or a cancelled prompt does nothing.
func (e *editor) goToLine() {
	s, ok := tty.ReadLine(e.t, ":", nil, "")
	if !ok {
		return
	}
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return
	}
	e.place(pos{n - 1, 0})
}
