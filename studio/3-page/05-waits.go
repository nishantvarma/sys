// waits is what a slot's page view waits on: fm's and ed's, ed's lenses
// among it, are a step of the plan; a shell on a page runs under ttyd,
// which waits on a week's test of where the work is done. Each says
// where it is written.
func waits(cmd string) string {
	word, _, _ := strings.Cut(cmd, " ")
	switch word {
	case "fm", "ed":
		return "Its page view is a step of the plan:" +
			" todo/03-studio/04-page."
	case "":
		return "A shell on a page runs under ttyd, which waits on a" +
			" week's test:" +
			" todo/03-studio/07-terminal:/for a week, note why/."
	}
	return "It runs in the terminal, and has no page view."
}
