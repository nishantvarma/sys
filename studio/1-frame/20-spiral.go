// spiral is slots laid out across, when across, else down: the first takes
// the golden part, at the right or at the top, and the rest go on the
// other way in what remains, at the left or at the bottom.
func spiral(slots []*slot, across bool) []*slot {
	first, rest := slots[0], slots[1:]
	other := rest[0]
	if len(rest) > 1 {
		other = &slot{kids: spiral(rest, !across)}
	}
	first.share, other.share = golden, 100-golden
	if across {
		return []*slot{other, first}
	}
	return []*slot{first, other}
}
