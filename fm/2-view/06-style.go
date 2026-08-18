func style(f entry) (func(string) string, string) {
	switch {
	case f.bad:
		return tty.Red, "@"
	case f.link:
		return tty.Cyan, "@"
	case f.dir:
		return tty.Blue, "/"
	case f.exec:
		return tty.Green, "*"
	default:
		return tty.Plain, ""
	}
}
