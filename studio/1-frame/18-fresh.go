// fresh checks a slot's or a tab's name as the frame is read: one word
// (word), none of the names studio keeps its own by, frame, face and here
// (open), and none that a slot or a tab before it took, for a name is how
// a program finds a pane.
func fresh(at, name string, named map[string]bool) error {
	why := ""
	switch {
	case !word(name):
		why = "a name is one word: letters, digits, - and _"
	case name == "frame" || name == "face" || name == "here":
		why = "studio keeps its own by that name"
	case named[name]:
		why = "a second slot or tab by that name"
	}
	if why != "" {
		return fmt.Errorf("%s: %s: %s", at, name, why)
	}
	named[name] = true
	return nil
}
