func modeColour(name string) string {
	if name == "insert" {
		return tty.Green(name)
	}
	return tty.Blue(name)
}
