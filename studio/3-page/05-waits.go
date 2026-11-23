// waits is what a slot's page view waits on: fm's and ed's are a step of
// the plan; the lens follows cur, which a page reads only once fm draws
// there; a shell on a page runs under ttyd, which waits on a week's test
// of where the work is done. Each says where it is written.
func waits(cmd string) string {
	word, _, _ := strings.Cut(cmd, " ")
	switch word {
	case "fm", "ed":
		return "Its page view is a step of the plan:" +
			" todo/studio/outline:/a page view of fm and ed/."
	case "studio":
		return "The lens follows cur in the terminal; on a page it" +
			" waits on fm's page view."
	case "":
		return "A shell on a page runs under ttyd, which waits on a" +
			" week's test:" +
			" todo/studio/terminal:/for a week, note why/."
	}
	return "It runs in the terminal, and has no page view."
}
