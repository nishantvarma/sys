// tally is what a split's shares give together, and how many of its
// slots give none, and so take an even part of what the rest leave.
func tally(kids []*slot) (given, none int) {
	for _, k := range kids {
		given += k.share
		if k.share == 0 {
			none++
		}
	}
	return given, none
}
