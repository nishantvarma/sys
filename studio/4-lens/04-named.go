// lensNamed finds a lens by its name, or by its number from 1, and its
// place among the lenses.
func lensNamed(s string) (lens, int, bool) {
	for i, l := range lenses {
		if l.name == s || strconv.Itoa(i+1) == s {
			return l, i, true
		}
	}
	return lens{}, 0, false
}
