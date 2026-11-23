// room checks a split's shares as the frame is read: together they may
// not pass the whole, and they must leave something for each slot that
// names none.
func room(kids []*slot, at string) error {
	given, none := tally(kids)
	switch {
	case given > 100:
		return fmt.Errorf("%s: its slots' shares come to %d%%, past"+
			" the whole", at, given)
	case given == 100 && none > 0:
		return fmt.Errorf("%s: its shares leave nothing for %d more",
			at, none)
	}
	return nil
}
