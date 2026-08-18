// revealAddr lands the cursor on an address: path, or path:line with the
// line riding along for the next edit. rg's path:line:text splits the same
// way, losing only the text. f stays outside — it yields a bare path, free
// to carry a colon.
func (m *fm) revealAddr(addr string) {
	if addr == "" {
		return
	}
	p, rest, _ := strings.Cut(addr, ":")
	num, _, _ := strings.Cut(rest, ":")
	n, _ := strconv.Atoi(num)
	m.reveal(p, n)
}
