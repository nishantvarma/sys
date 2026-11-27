// fresh checks a slot's or a tab's name as the frame is read: one word
// (word), and none that a slot or a tab before it took, for a name is how
// a program finds a pane.
func fresh(at, name string, named map[string]bool) error {
	why := ""
	switch {
	case !word(name):
		why = "a name is one word: letters, digits, - and _"
	case named[name]:
		why = "a second slot or tab by that name"
	}
	if why != "" {
		return fmt.Errorf("%s: %s: %s", at, name, why)
	}
	named[name] = true
	return nil
}
