// word is whether a slot's name is one word, as a program says it on a
// command line and tmux keeps it in an option's name: letters, digits, -
// and _.
func word(s string) bool {
	return s != "" && !strings.ContainsFunc(s, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r) &&
			r != '-' && r != '_'
	})
}
