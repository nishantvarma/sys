// fzmemory does a line of the log again: a line is a verb then an address,
// so the pick is a sentence and running it is that act repeated. the address
// is whole, so the verb carries its own place and fm does not move — going
// there was never what the pick asked for. a line whose verb is c is the one
// exception: c was a select, and a select done again is the landing.
func (m *fm) fzmemory() {
	verb, addr, ok := strings.Cut(m.pick("m"), " ")
	if !ok {
		return
	}
	if verb == "c" {
		// the select is the landing; there is nothing to run
		m.revealAddr(addr)
		return
	}
	m.spawn(false, false, verb, addr)
}
