// next is the live tab after the one at at, round, of a slot's tabs tps,
// to take that one's place (died, closeTab); false when none is left.
func next(tps []tabPane, at int) (tabPane, bool) {
	for i := 1; i < len(tps); i++ {
		if n := tps[(at+i)%len(tps)]; !n.dead {
			return n, true
		}
	}
	return tabPane{}, false
}
