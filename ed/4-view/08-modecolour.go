// modeColour is the colour the status says a mode in, from tty's palette:
// insert as an addition is, extend as a link is, normal in the accent.
func modeColour(name string) string {
	switch name {
	case "insert":
		return tty.Plus
	case "extend":
		return tty.Link
	}
	return tty.Accent
}
