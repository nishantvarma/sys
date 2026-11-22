// look is how the terminal draws a kind: its colour, and the mark after
// its name.
func look(kind string) (func(string) string, string) {
	switch kind {
	case "bad":
		return tty.Red, "@"
	case "link":
		return tty.Cyan, "@"
	case "dir":
		return tty.Blue, "/"
	case "exec":
		return tty.Green, "*"
	default:
		return tty.Plain, ""
	}
}
